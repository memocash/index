package slp

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Baton struct {
	TxHash    [32]byte
	Index     uint32
	TokenHash [32]byte
}

func (o *Baton) GetTopic() string {
	return db.TopicSlpBaton
}

func (o *Baton) GetShardSource() uint {
	return client.GenShardSource(o.TxHash[:])
}

func (o *Baton) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(o.TxHash[:]),
		jutil.GetUint32Data(o.Index),
	)
}

func (o *Baton) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+4 {
		return
	}
	copy(o.TxHash[:32], jutil.ByteReverse(uid[:32]))
	o.Index = jutil.GetUint32(uid[32:])
}

func (o *Baton) Serialize() []byte {
	return jutil.ByteReverse(o.TokenHash[:])
}

func (o *Baton) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength {
		return
	}
	copy(o.TokenHash[:], jutil.ByteReverse(data))
}

func GetBatons(ctx context.Context, outs []memo.Out) ([]*Baton, error) {
	var shardUids = make(map[uint32][][]byte)
	for _, out := range outs {
		shard := db.GetShardIdFromByte32(out.TxHash)
		shardUids[shard] = append(shardUids[shard], jutil.CombineBytes(
			jutil.ByteReverse(out.TxHash[:]),
			jutil.GetUint32Data(out.Index),
		))
	}
	messages, err := db.GetSpecific(ctx, db.TopicSlpBaton, shardUids)
	if err != nil {
		return nil, jerr.Get("error getting slp batons", err)
	}
	var batons []*Baton
	for i := range messages {
		var baton = new(Baton)
		db.Set(baton, messages[i])
		batons = append(batons, baton)
	}
	return batons, nil
}
