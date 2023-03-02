package slp

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type Baton struct {
	TxHash    [32]byte
	Index     uint32
	TokenHash [32]byte
}

func (o *Baton) GetTopic() string {
	return db.TopicSlpBaton
}

func (o *Baton) GetShard() uint {
	return client.GetByteShard(o.TxHash[:])
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
	copy(o.TxHash[:32], jutil.ByteReverse(uid))
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

func GetBaton(txHash [32]byte, index uint32) (*Baton, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(txHash[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	uid := jutil.CombineBytes(jutil.ByteReverse(txHash[:]), jutil.GetUint32Data(index))
	if err := dbClient.GetSingle(db.TopicSlpBaton, uid); err != nil {
		return nil, jerr.Get("error getting client message slp baton", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, jerr.Get("error no messages slp baton", client.EntryNotFoundError)
	} else if len(dbClient.Messages) > 1 {
		return nil, jerr.Newf("error unexpected number of messages slp batons: %d", len(dbClient.Messages))
	}
	var slpBaton = new(Baton)
	db.Set(slpBaton, dbClient.Messages[0])
	return slpBaton, nil
}
