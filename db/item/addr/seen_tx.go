package addr

import (
	"context"
	"fmt"
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

func (i *SeenTx) GetShardSource() uint {
	return client.GenShardSource(i.Addr[:])
}

func (i *SeenTx) GetUid() []byte {
	return jutil.CombineBytes(
		i.Addr[:],
		jutil.GetTimeByteNanoBig(i.Seen),
		jutil.ByteReverse(i.TxHash[:]),
	)
}

func (i *SeenTx) SetUid(uid []byte) {
	if len(uid) != 65 {
		return
	}
	copy(i.Addr[:], uid[:25])
	i.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(i.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (i *SeenTx) Serialize() []byte {
	return nil
}

func (i *SeenTx) Deserialize([]byte) {}

func GetSeenTxs(ctx context.Context, addr [25]byte, start []byte) ([]*SeenTx, error) {
	shardConfig := config.GetShardConfig(client.GenShardSource32(addr[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Context:  ctx,
		Topic:    db.TopicAddrSeenTx,
		Start:    start,
		Prefixes: [][]byte{addr[:]},
		Max:      client.ExLargeLimit,
	}); err != nil {
		return nil, jerr.Get("error getting db addr seen txs by prefix", err)
	}
	var seenTxs = make([]*SeenTx, len(dbClient.Messages))
	for i := range dbClient.Messages {
		seenTxs[i] = new(SeenTx)
		db.Set(seenTxs[i], dbClient.Messages[i])
	}
	return seenTxs, nil
}

func ListenAddrSeenTxs(ctx context.Context, addrs [][25]byte) (chan *SeenTx, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range addrs {
		shard := db.GetShardIdFromByte32(addrs[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addrs[i][:])
	}
	chanMessages, err := db.ListenPrefixes(ctx, db.TopicAddrSeenTx, shardPrefixes)
	if err != nil {
		return nil, fmt.Errorf("error getting listen prefixes for address seen tx; %w", err)
	}
	var chanSeenTxs = make(chan *SeenTx)
	go func() {
		for {
			msg, ok := <-chanMessages
			if !ok {
				return
			}
			var addrSeenTx = new(SeenTx)
			db.Set(addrSeenTx, *msg)
			chanSeenTxs <- addrSeenTx
		}
	}()
	return chanSeenTxs, nil
}
