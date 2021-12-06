package item

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
	"sort"
	"sync"
)

type BlockTxRaw struct {
	BlockHash []byte
	TxHash    []byte
	Raw       []byte
}

func (t BlockTxRaw) GetUid() []byte {
	return GetBlockTxRawUid(t.BlockHash, t.TxHash)
}

func (t BlockTxRaw) GetShard() uint {
	return client.GetByteShard(t.TxHash)
}

func (t BlockTxRaw) GetTopic() string {
	return TopicBlockTxRaw
}

func (t BlockTxRaw) Serialize() []byte {
	return t.Raw
}

func (t *BlockTxRaw) SetUid(uid []byte) {
	if len(uid) != 64 {
		return
	}
	t.BlockHash = jutil.ByteReverse(uid[:32])
	t.TxHash = jutil.ByteReverse(uid[32:64])
}

func (t *BlockTxRaw) Deserialize(data []byte) {
	t.Raw = data
}

func GetBlockTxRawUid(blockHash, txHash []byte) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(blockHash), jutil.ByteReverse(txHash))
}

func GetRawBlockTxByHash(blockHash, txHash []byte) (*BlockTxRaw, error) {
	shardConfig := config.GetShardConfig(GetShardByte32(txHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	err := dbClient.GetSingle(TopicBlockTxRaw, GetBlockTxRawUid(blockHash, txHash))
	if err != nil {
		return nil, jerr.Get("error getting client message raw tx by hash", err)
	}
	if len(dbClient.Messages) != 1 {
		return nil, jerr.Newf("error unexpected number of client messages raw tx by hash returned (%d)",
			len(dbClient.Messages))
	}
	var tx = new(BlockTxRaw)
	tx.SetUid(dbClient.Messages[0].Uid)
	tx.Deserialize(dbClient.Messages[0].Message)
	return tx, nil
}

func GetRawTxBlocksByHashes(txBlocks []*TxBlock) ([]*BlockTxRaw, error) {
	var blockTxes = make([]*BlockTx, len(txBlocks))
	for i := range blockTxes {
		blockTxes[i] = &BlockTx{
			TxHash:    txBlocks[i].TxHash,
			BlockHash: txBlocks[i].BlockHash,
		}
	}
	blockTxRaw, err := GetRawBlockTxsByHashes(blockTxes)
	if err != nil {
		return nil, jerr.Get("error getting raw block txs by hashes", err)
	}
	return blockTxRaw, nil
}

func GetRawBlockTxsByTxHashes(blockHash []byte, txHashes [][]byte) ([]*BlockTxRaw, error) {
	var blockTxs = make([]*BlockTx, len(txHashes))
	for i := range txHashes {
		blockTxs[i] = &BlockTx{
			TxHash:    txHashes[i],
			BlockHash: blockHash,
		}
	}
	blockTxRaws, err := GetRawBlockTxsByHashes(blockTxs)
	if err != nil {
		return nil, jerr.Get("error getting raw block txs", err)
	}
	return blockTxRaws, nil
}

func GetRawBlockTxsByHashes(blockTxs []*BlockTx) ([]*BlockTxRaw, error) {
	var shardUids = make(map[uint32][][]byte)
	for _, blockTx := range blockTxs {
		shard := GetShardByte32(blockTx.TxHash)
		shardUids[shard] = append(shardUids[shard], GetBlockTxRawUid(blockTx.BlockHash, blockTx.TxHash))
	}
	var shardTxs = make(map[uint32][]*BlockTxRaw)
	var wg sync.WaitGroup
	wg.Add(len(shardUids))
	var lock = sync.RWMutex{}
	var errs []error
	for shardT, uidsT := range shardUids {
		go func(shard uint32, uids [][]byte) {
			defer wg.Done()
			shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
			dbClient := client.NewClient(shardConfig.GetHost())
			err := dbClient.GetSpecific(TopicBlockTxRaw, uids)
			if err != nil {
				errs = append(errs, jerr.Get("error getting client raw tx message", err))
				return
			}
			for _, msg := range dbClient.Messages {
				var tx = new(BlockTxRaw)
				tx.SetUid(msg.Uid)
				tx.Deserialize(msg.Message)
				lock.Lock()
				shardTxs[shard] = append(shardTxs[shard], tx)
				lock.Unlock()
			}
		}(shardT, uidsT)
	}
	wg.Wait()
	if len(errs) > 0 {
		return nil, jerr.Get("error getting raw tx messages", jerr.Combine(errs...))
	}
	var allTxs []*BlockTxRaw
	for _, txs := range shardTxs {
		allTxs = append(allTxs, txs...)
	}
	return allTxs, nil
}

func GetRawBlockTxs(shard uint32, offset uint64) ([]*BlockTxRaw, error) {
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	err := dbClient.GetLarge(TopicTx, nil, true, false)
	if err != nil {
		return nil, jerr.Get("error getting client message", err)
	}
	var txs = make([]*BlockTxRaw, len(dbClient.Messages))
	for i := range dbClient.Messages {
		txs[i] = new(BlockTxRaw)
		txs[i].SetUid(dbClient.Messages[i].Uid)
		txs[i].Deserialize(dbClient.Messages[i].Message)
	}
	return txs, nil
}

type BlockTxesRawRequest struct {
	Shard       uint32
	BlockHash   []byte
	StartTxHash []byte
	Limit       uint32
	Wait        bool
}

func (r BlockTxesRawRequest) GetStartUid() []byte {
	return GetBlockTxRawUid(r.BlockHash, r.StartTxHash)
}

func (r BlockTxesRawRequest) GetStartUidPlusOne() []byte {
	return client.IncrementBytes(r.GetStartUid())
}

func GetBlockTxesRaw(request BlockTxesRawRequest) ([]*BlockTxRaw, error) {
	limit := request.Limit
	if limit == 0 {
		limit = client.LargeLimit
	}
	shardConfig := config.GetShardConfig(request.Shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	err := dbClient.GetWOpts(client.Opts{
		Topic:    TopicBlockTxRaw,
		Prefixes: [][]byte{jutil.ByteReverse(request.BlockHash)},
		Start:    request.GetStartUid(),
		Max:      limit,
		Wait:     request.Wait,
	})
	if err != nil {
		return nil, jerr.Get("error getting block txes raw client message", err)
	}
	var blockTxRaws = make([]*BlockTxRaw, len(dbClient.Messages))
	for i := range dbClient.Messages {
		blockTxRaws[i] = new(BlockTxRaw)
		blockTxRaws[i].SetUid(dbClient.Messages[i].Uid)
		blockTxRaws[i].Deserialize(dbClient.Messages[i].Message)
	}
	sort.Slice(blockTxRaws, func(i, j int) bool {
		return bytes.Compare(blockTxRaws[i].BlockHash, blockTxRaws[j].BlockHash) == -1
	})
	return blockTxRaws, nil
}
