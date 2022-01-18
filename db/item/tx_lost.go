package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
)

type TxLost struct {
	TxHash      []byte
	DoubleSpend []byte
}

func (l TxLost) GetUid() []byte {
	return jutil.CombineBytes(jutil.ByteReverse(l.TxHash), jutil.ByteReverse(l.DoubleSpend))
}

func (l TxLost) GetShard() uint {
	return client.GetByteShard(l.TxHash)
}

func (l TxLost) GetTopic() string {
	return TopicTxLost
}

func (l TxLost) Serialize() []byte {
	return nil
}

func (l *TxLost) SetUid(uid []byte) {
	if len(uid) != 32 && len(uid) != 64 {
		return
	}
	l.TxHash = jutil.ByteReverse(uid[:32])
	if len(uid) != 64 {
		return
	}
	l.DoubleSpend = jutil.ByteReverse(uid[32:])
}

func (l *TxLost) Deserialize([]byte) {}

func GetTxLosts(txHashes [][]byte) ([]*TxLost, error) {
	var shardTxHashGroups = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := GetShardByte32(txHash)
		shardTxHashGroups[shard] = append(shardTxHashGroups[shard], txHash)
	}
	var txLosts []*TxLost
	for shard, groupTxHashes := range shardTxHashGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		var prefixes = make([][]byte, len(groupTxHashes))
		for i := range groupTxHashes {
			prefixes[i] = jutil.ByteReverse(groupTxHashes[i])
		}
		if err := db.GetByPrefixes(TopicTxLost, prefixes); err != nil {
			return nil, jerr.Get("error getting by prefixes for tx losts", err)
		}
		for i := range db.Messages {
			var txLost = new(TxLost)
			txLost.SetUid(db.Messages[i].Uid)
			txLosts = append(txLosts, txLost)
		}
	}
	return txLosts, nil
}

func GetAllTxLosts(shard uint32, startTxLost []byte) ([]*TxLost, error) {
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	var txLosts []*TxLost
	if err := db.GetWOpts(client.Opts{
		Topic: TopicTxLost,
		Start: jutil.ByteReverse(startTxLost),
		Max:   client.DefaultLimit,
	}); err != nil {
		return nil, jerr.Get("error getting all tx losts", err)
	}
	for i := range db.Messages {
		var txLost = new(TxLost)
		txLost.SetUid(db.Messages[i].Uid)
		txLosts = append(txLosts, txLost)
	}
	return txLosts, nil
}

func RemoveTxLosts(txLosts []*TxLost) error {
	var shardUidsMap = make(map[uint32][][]byte)
	for _, txLost := range txLosts {
		shard := uint32(txLost.GetShard())
		shardUidsMap[shard] = append(shardUidsMap[shard], txLost.GetUid())
	}
	for shard, shardUids := range shardUidsMap {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		if err := db.DeleteMessages(TopicTxLost, shardUids); err != nil {
			return jerr.Get("error deleting topic tx losts", err)
		}
	}
	return nil
}
