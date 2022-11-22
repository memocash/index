package chain

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
	"sort"
)

type OutputInput struct {
	PrevHash  [32]byte
	PrevIndex uint32
	Hash      [32]byte
	Index     uint32
}

func (t *OutputInput) GetTopic() string {
	return db.TopicChainOutputInput
}

func (t *OutputInput) GetShard() uint {
	return client.GetByteShard(t.PrevHash[:])
}

func (t *OutputInput) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(t.PrevHash[:]),
		jutil.GetUint32DataBig(t.PrevIndex),
		jutil.ByteReverse(t.Hash[:]),
		jutil.GetUint32DataBig(t.Index),
	)
}

func (t *OutputInput) SetUid(uid []byte) {
	if len(uid) != 72 {
		return
	}
	copy(t.PrevHash[:], jutil.ByteReverse(uid[:32]))
	t.PrevIndex = jutil.GetUint32Big(uid[32:36])
	copy(t.Hash[:], jutil.ByteReverse(uid[36:68]))
	t.Index = jutil.GetUint32Big(uid[68:72])
}

func (t *OutputInput) Serialize() []byte {
	return nil
}

func (t *OutputInput) Deserialize([]byte) {}

func GetOutputInput(out memo.Out) ([]*OutputInput, error) {
	shard := db.GetShardByte32(out.TxHash)
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	prefix := jutil.CombineBytes(jutil.ByteReverse(out.TxHash), jutil.GetUint32Data(out.Index))
	if err := dbClient.GetByPrefix(db.TopicChainOutputInput, prefix); err != nil {
		return nil, jerr.Get("error getting by prefix for chain output input", err)
	}
	var outputInputs = make([]*OutputInput, len(dbClient.Messages))
	for i := range dbClient.Messages {
		outputInputs[i] = new(OutputInput)
		db.Set(outputInputs[i], dbClient.Messages[i])
	}
	return outputInputs, nil
}

func GetOutputInputs(outs []memo.Out) ([]*OutputInput, error) {
	var shardOutGroups = make(map[uint32][]memo.Out)
	for _, out := range outs {
		shard := db.GetShardByte32(out.TxHash)
		shardOutGroups[shard] = append(shardOutGroups[shard], out)
	}
	wait := db.NewWait(len(shardOutGroups))
	var outputInputs []*OutputInput
	for shardT, outGroupT := range shardOutGroups {
		go func(shard uint32, outGroup []memo.Out) {
			defer wait.Group.Done()
			shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
			dbClient := client.NewClient(shardConfig.GetHost())
			var prefixes = make([][]byte, len(outGroup))
			for i := range outGroup {
				prefixes[i] = jutil.CombineBytes(
					jutil.ByteReverse(outGroup[i].TxHash),
					jutil.GetUint32Data(outGroup[i].Index),
				)
			}
			sort.Slice(prefixes, func(i, j int) bool {
				return jutil.ByteLT(prefixes[i], prefixes[j])
			})
			for len(prefixes) > 0 {
				var prefixesToUse [][]byte
				if len(prefixes) > client.HugeLimit {
					prefixesToUse, prefixes = prefixes[:client.HugeLimit], prefixes[client.HugeLimit:]
				} else {
					prefixesToUse, prefixes = prefixes, nil
				}
				if err := dbClient.GetByPrefixes(db.TopicChainOutputInput, prefixesToUse); err != nil {
					wait.AddError(jerr.Get("error getting by prefixes for chain output inputs", err))
					return
				}
				wait.Lock.Lock()
				for i := range dbClient.Messages {
					var outputInput = new(OutputInput)
					db.Set(outputInput, dbClient.Messages[i])
					outputInputs = append(outputInputs, outputInput)
				}
				wait.Lock.Unlock()
			}
		}(shardT, outGroupT)
	}
	wait.Group.Wait()
	if len(wait.Errs) > 0 {
		return nil, jerr.Get("error getting chain output input messages", jerr.Combine(wait.Errs...))
	}
	return outputInputs, nil
}

func GetOutputInputsForTxHashes(txHashes [][]byte) ([]*OutputInput, error) {
	var shardOutGroups = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := db.GetShardByte32(txHash)
		shardOutGroups[shard] = append(shardOutGroups[shard], txHash)
	}
	var outputInputs []*OutputInput
	for shard, outGroup := range shardOutGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		var prefixes = make([][]byte, len(outGroup))
		for i := range outGroup {
			prefixes[i] = jutil.ByteReverse(outGroup[i])
		}
		if err := dbClient.GetByPrefixes(db.TopicChainOutputInput, prefixes); err != nil {
			return nil, jerr.Get("error getting by prefixes for chain output inputs by tx hashes", err)
		}
		for i := range dbClient.Messages {
			var outputInput = new(OutputInput)
			db.Set(outputInput, dbClient.Messages[i])
			outputInputs = append(outputInputs, outputInput)
		}
	}
	return outputInputs, nil
}
