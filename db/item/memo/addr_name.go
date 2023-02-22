package memo

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
	"time"
)

type AddrName struct {
	Addr   [25]byte
	Seen   time.Time
	TxHash [32]byte
	Name   string
}

func (n *AddrName) GetUid() []byte {
	return jutil.CombineBytes(
		n.Addr[:],
		jutil.GetTimeByteNanoBig(n.Seen),
		jutil.ByteReverse(n.TxHash[:]),
	)
}

func (n *AddrName) GetShard() uint {
	return client.GetByteShard(n.Addr[:])
}

func (n *AddrName) GetTopic() string {
	return db.TopicMemoAddrName
}

func (n *AddrName) Serialize() []byte {
	return []byte(n.Name)
}

func (n *AddrName) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(n.Addr[:], uid[:25])
	n.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(n.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (n *AddrName) Deserialize(data []byte) {
	n.Name = string(data)
}

func GetAddrName(ctx context.Context, addr [25]byte) (*AddrName, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(addr[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoAddrName,
		Prefixes: [][]byte{addr[:]},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db addr memo name by prefix", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, jerr.Get("error no addr memo names found", client.EntryNotFoundError)
	}
	var addrName = new(AddrName)
	db.Set(addrName, dbClient.Messages[0])
	return addrName, nil
}

func ListenAddrNames(ctx context.Context, addrs [][25]byte) (chan *AddrName, error) {
	if len(addrs) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, addr := range addrs {
		shard := client.GetByteShard32(addr[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addr[:])
	}
	shardConfigs := config.GetQueueShards()
	var addrNameChan = make(chan *AddrName)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(addrNameChan)
	})
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoAddrName, prefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db addr memo names by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var addrName = new(AddrName)
				db.Set(addrName, *msg)
				addrNameChan <- addrName
			}
			cancelCtx.Cancel()
		}()
	}
	return addrNameChan, nil
}
