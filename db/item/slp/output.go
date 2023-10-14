package slp

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
	"sort"
)

type Output struct {
	TxHash    [32]byte
	Index     uint32
	TokenHash [32]byte
	Quantity  uint64
}

func (o *Output) GetTopic() string {
	return db.TopicSlpOutput
}

func (o *Output) GetShard() uint {
	return client.GetByteShard(o.TxHash[:])
}

func (o *Output) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(o.TxHash[:]),
		jutil.GetUint32Data(o.Index),
	)
}

func (o *Output) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+4 {
		return
	}
	copy(o.TxHash[:32], jutil.ByteReverse(uid))
	o.Index = jutil.GetUint32(uid[32:])
}

func (o *Output) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(o.TokenHash[:]),
		jutil.GetUint64Data(o.Quantity),
	)
}

func (o *Output) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength+8 {
		return
	}
	copy(o.TokenHash[:], jutil.ByteReverse(data[:32]))
	o.Quantity = jutil.GetUint64(data[32:])
}

func GetOutput(txHash [32]byte, index uint32) (*Output, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(txHash[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	uid := jutil.CombineBytes(jutil.ByteReverse(txHash[:]), jutil.GetUint32Data(index))
	if err := dbClient.GetSingle(db.TopicSlpOutput, uid); err != nil {
		return nil, jerr.Get("error getting client message slp output", err)
	}
	if len(dbClient.Messages) != 1 {
		return nil, jerr.Newf("error unexpected number of messages slp outputs: %d", len(dbClient.Messages))
	}
	var slpOutput = new(Output)
	db.Set(slpOutput, dbClient.Messages[0])
	return slpOutput, nil
}

func GetOutputs(outs []memo.Out) ([]*Output, error) {
	var shardOutGroups = make(map[uint32][]memo.Out)
	for _, out := range outs {
		shard := db.GetShardByte32(out.TxHash)
		shardOutGroups[shard] = append(shardOutGroups[shard], out)
	}
	wait := db.NewWait(len(shardOutGroups))
	var outputs []*Output
	for shardT, outGroupT := range shardOutGroups {
		go func(shard uint32, outGroup []memo.Out) {
			defer wait.Group.Done()
			shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
			dbClient := client.NewClient(shardConfig.GetHost())
			var uids = make([][]byte, len(outGroup))
			for i := range outGroup {
				txHash, err := chainhash.NewHash(outGroup[i].TxHash)
				if err != nil {
					wait.AddError(jerr.Get("error creating tx hash", err))
				}
				uids[i] = db.GetTxHashIndexUid(txHash[:], outGroup[i].Index)
			}
			sort.Slice(uids, func(i, j int) bool {
				return jutil.ByteLT(uids[i], uids[j])
			})
			if err := dbClient.GetSpecific(db.TopicSlpOutput, uids); err != nil {
				wait.AddError(jerr.Get("error getting specific slp outputs by uids", err))
				return
			}
			wait.Lock.Lock()
			for i := range dbClient.Messages {
				var output = new(Output)
				db.Set(output, dbClient.Messages[i])
				outputs = append(outputs, output)
			}
			wait.Lock.Unlock()
		}(shardT, outGroupT)
	}
	wait.Group.Wait()
	if len(wait.Errs) > 0 {
		return nil, jerr.Get("error getting slp outputs", jerr.Combine(wait.Errs...))
	}
	return outputs, nil
}
