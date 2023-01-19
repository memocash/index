package addr

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
	"time"
)

type SeenTx struct {
	Addr   [25]byte
	Seen   time.Time
	TxHash [32]byte
}

func (i *SeenTx) GetTopic() string {
	return db.TopicAddrSeenTx
}

func (i *SeenTx) GetShard() uint {
	return client.GetByteShard(i.Addr[:])
}

func (i *SeenTx) GetUid() []byte {
	return jutil.CombineBytes(
		i.Addr[:],
		jutil.GetTimeByte(i.Seen),
		jutil.ByteReverse(i.TxHash[:]),
	)
}

func (i *SeenTx) SetUid(uid []byte) {
	if len(uid) != 65 {
		return
	}
	copy(i.Addr[:], uid[:25])
	i.Seen = jutil.GetByteTime(uid[25:33])
	copy(i.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (i *SeenTx) Serialize() []byte {
	return nil
}

func (i *SeenTx) Deserialize([]byte) {}

func GetSeenTxs(addr [25]byte, start []byte) ([]*SeenTx, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(addr[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicAddrSeenTx,
		Start:    start,
		Prefixes: [][]byte{addr[:]},
		Max:      client.ExLargeLimit,
	}); err != nil {
		return nil, jerr.Get("error getting db addr seen txs by prefix", err)
	}
	var heightInputs = make([]*SeenTx, len(dbClient.Messages))
	for i := range dbClient.Messages {
		heightInputs[i] = new(SeenTx)
		db.Set(heightInputs[i], dbClient.Messages[i])
	}
	return heightInputs, nil
}

func ListenMempoolAddrSeenTxsMultiple(ctx context.Context, addrs [][25]byte) ([]chan *SeenTx, error) {
	var shardAddrGroups = make(map[uint32][][]byte)
	for _, addr := range addrs {
		shard := db.GetShardByte32(addr[:])
		shardAddrGroups[shard] = append(shardAddrGroups[shard], addr[:])
	}
	var chanSeenTxs []chan *SeenTx
	for shard, addrGroup := range shardAddrGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(ctx, db.TopicAddrSeenTx, addrGroup)
		if err != nil {
			return nil, jerr.Get("error getting addr seen txs listen message chan", err)
		}
		var chanAddrSeenTx = make(chan *SeenTx)
		go func() {
			for {
				msg, ok := <-chanMessage
				if !ok {
					close(chanAddrSeenTx)
					return
				}
				var addrSeenTx = new(SeenTx)
				db.Set(addrSeenTx, *msg)
				chanAddrSeenTx <- addrSeenTx
			}
		}()
		chanSeenTxs = append(chanSeenTxs, chanAddrSeenTx)
	}
	return chanSeenTxs, nil
}
