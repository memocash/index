package chain

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type TxInput struct {
	TxHash       [32]byte
	Index        uint32
	PrevHash     [32]byte
	PrevIndex    uint32
	Sequence     uint32
	UnlockScript []byte
}

func (t *TxInput) GetTopic() string {
	return db.TopicChainTxInput
}

func (t *TxInput) GetShardSource() uint {
	return client.GenShardSource(t.TxHash[:])
}

func (t *TxInput) GetUid() []byte {
	return GetTxInputUid(t.TxHash, t.Index)
}

func (t *TxInput) SetUid(uid []byte) {
	if len(uid) != 36 {
		return
	}
	copy(t.TxHash[:], jutil.ByteReverse(uid[:32]))
	t.Index = jutil.GetUint32Big(uid[32:36])
}

func (t *TxInput) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(t.PrevHash[:]),
		jutil.GetUint32DataBig(t.PrevIndex),
		jutil.GetUint32Data(t.Sequence),
		t.UnlockScript,
	)
}

func (t *TxInput) Deserialize(data []byte) {
	if len(data) < 40 {
		return
	}
	copy(t.PrevHash[:], jutil.ByteReverse(data[:32]))
	t.PrevIndex = jutil.GetUint32Big(data[32:36])
	t.Sequence = jutil.GetUint32(data[36:40])
	t.UnlockScript = data[40:]
}

func GetAllTxInputs(shard uint32, startUid []byte) ([]*TxInput, error) {
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic: db.TopicChainTxInput,
		Start: startUid,
		Max:   client.HugeLimit,
	}); err != nil {
		return nil, jerr.Get("error getting db message chain tx inputs for all", err)
	}
	var txInputs = make([]*TxInput, len(dbClient.Messages))
	for i := range dbClient.Messages {
		txInputs[i] = new(TxInput)
		db.Set(txInputs[i], dbClient.Messages[i])
	}
	return txInputs, nil
}

func GetTxInputUid(txHash [32]byte, index uint32) []byte {
	return db.GetTxHashIndexUid(txHash[:], index)
}

func GetTxInputsByHashes(ctx context.Context, txHashes [][32]byte) ([]*TxInput, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range txHashes {
		shard := uint32(db.GetShardIdFromByte(txHashes[i][:]))
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(txHashes[i][:]))
	}
	messages, err := db.GetByPrefixes(ctx, db.TopicChainTxInput, shardPrefixes)
	if err != nil {
		return nil, jerr.Get("error getting client message chain tx input", err)
	}
	var txInputs []*TxInput
	for _, msg := range messages {
		var txInput = new(TxInput)
		db.Set(txInput, msg)
		txInputs = append(txInputs, txInput)
	}
	return txInputs, nil
}

func GetTxInputs(ctx context.Context, outs []memo.Out) ([]*TxInput, error) {
	var shardUids = make(map[uint32][][]byte)
	for _, out := range outs {
		shard := db.GetShardIdFromByte32(out.TxHash)
		shardUids[shard] = append(shardUids[shard], GetTxInputUid(db.RawTxHashToFixed(out.TxHash), out.Index))
	}
	messages, err := db.GetSpecific(ctx, db.TopicChainTxInput, shardUids)
	if err != nil {
		return nil, jerr.Get("error getting client message chain tx input", err)
	}
	var txInputs []*TxInput
	for i := range messages {
		var txInput = new(TxInput)
		db.Set(txInput, messages[i])
		txInputs = append(txInputs, txInput)
	}
	return txInputs, nil
}
