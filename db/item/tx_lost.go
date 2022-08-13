package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
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
	return db.TopicTxLost
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
		shard := db.GetShardByte32(txHash)
		shardTxHashGroups[shard] = append(shardTxHashGroups[shard], txHash)
	}
	var txLosts []*TxLost
	for shard, groupTxHashes := range shardTxHashGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		var prefixes = make([][]byte, len(groupTxHashes))
		for i := range groupTxHashes {
			prefixes[i] = jutil.ByteReverse(groupTxHashes[i])
		}
		if err := dbClient.GetByPrefixes(db.TopicTxLost, prefixes); err != nil {
			return nil, jerr.Get("error getting by prefixes for tx losts", err)
		}
		for i := range dbClient.Messages {
			var txLost = new(TxLost)
			db.Set(txLost, dbClient.Messages[i])
			txLosts = append(txLosts, txLost)
		}
	}
	return txLosts, nil
}

func GetAllTxLosts(shard uint32, lastUid []byte) ([]*TxLost, error) {
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var txLosts []*TxLost
	if err := dbClient.GetWOpts(client.Opts{
		Topic: db.TopicTxLost,
		Start: lastUid,
		Max:   client.HugeLimit,
	}); err != nil {
		return nil, jerr.Get("error getting all tx losts", err)
	}
	for i := range dbClient.Messages {
		var txLost = new(TxLost)
		db.Set(txLost, dbClient.Messages[i])
		txLosts = append(txLosts, txLost)
	}
	return txLosts, nil
}

func RemoveTxLosts(txLosts []*TxLost) error {
	var shardUidsMap = make(map[uint32][][]byte)
	for _, txLost := range txLosts {
		shard := uint32(db.GetShard(txLost.GetShard()))
		shardUidsMap[shard] = append(shardUidsMap[shard], txLost.GetUid())
	}
	for shard, shardUids := range shardUidsMap {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.DeleteMessages(db.TopicTxLost, shardUids); err != nil {
			return jerr.Get("error deleting topic tx losts", err)
		}
	}
	return nil
}
