package memo

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type AddrHeightName struct {
	Addr   [25]byte
	Height int64
	TxHash [32]byte
	Name   string
}

func (n *AddrHeightName) GetUid() []byte {
	return jutil.CombineBytes(
		n.Addr[:],
		jutil.ByteFlip(jutil.GetInt64DataBig(n.Height)),
		jutil.ByteReverse(n.TxHash[:]),
	)
}

func (n *AddrHeightName) GetShard() uint {
	return client.GetByteShard(n.Addr[:])
}

func (n *AddrHeightName) GetTopic() string {
	return db.TopicMemoAddrHeightName
}

func (n *AddrHeightName) Serialize() []byte {
	return []byte(n.Name)
}

func (n *AddrHeightName) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(n.Addr[:], uid[:25])
	n.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[25:33]))
	copy(n.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (n *AddrHeightName) Deserialize(data []byte) {
	n.Name = string(data)
}

func GetAddrHeightName(ctx context.Context, addr [25]byte) (*AddrHeightName, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(addr[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoAddrHeightName,
		Prefixes: [][]byte{addr[:]},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db addr memo name by prefix", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, jerr.Get("error no addr memo names found", client.EntryNotFoundError)
	}
	var addrName = new(AddrHeightName)
	db.Set(addrName, dbClient.Messages[0])
	return addrName, nil
}

func ListenAddrHeightNames(ctx context.Context, addrs [][25]byte) (chan *AddrHeightName, error) {
	if len(addrs) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, addr := range addrs {
		shard := client.GetByteShard32(addr[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addr[:])
	}
	shardConfigs := config.GetQueueShards()
	var addrNameChan = make(chan *AddrHeightName)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(addrNameChan)
	})
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoAddrHeightName, prefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db addr memo names by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var addrName = new(AddrHeightName)
				db.Set(addrName, *msg)
				addrNameChan <- addrName
			}
			cancelCtx.Cancel()
		}()
	}
	return addrNameChan, nil
}
