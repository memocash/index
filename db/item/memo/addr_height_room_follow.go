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

type AddrHeightRoomFollow struct {
	Addr     [25]byte
	Height   int64
	TxHash   [32]byte
	Unfollow bool
	Room     string
}

func (f *AddrHeightRoomFollow) GetTopic() string {
	return db.TopicMemoAddrHeightRoomFollow
}

func (f *AddrHeightRoomFollow) GetShard() uint {
	return client.GetByteShard(f.Addr[:])
}

func (f *AddrHeightRoomFollow) GetUid() []byte {
	return jutil.CombineBytes(
		f.Addr[:],
		jutil.ByteFlip(jutil.GetInt64DataBig(f.Height)),
		jutil.ByteReverse(f.TxHash[:]),
	)
}

func (f *AddrHeightRoomFollow) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(f.Addr[:], uid[:25])
	f.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[25:33]))
	copy(f.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (f *AddrHeightRoomFollow) Serialize() []byte {
	var unfollow byte
	if f.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		[]byte(f.Room),
	)
}

func (f *AddrHeightRoomFollow) Deserialize(data []byte) {
	if len(data) < 1 {
		return
	}
	f.Unfollow = data[0] == 1
	f.Room = string(data[1:])
}

func GetAddrHeightRoomFollows(ctx context.Context, addrs [][25]byte) ([]*AddrHeightRoomFollow, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, addr := range addrs {
		shard := client.GetByteShard32(addr[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addr[:])
	}
	shardConfigs := config.GetQueueShards()
	var addrFollows []*AddrHeightRoomFollow
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicMemoAddrHeightRoomFollow,
			Prefixes: prefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db memo addr room follow by prefix", err)
		}
		for _, msg := range dbClient.Messages {
			var addrFollow = new(AddrHeightRoomFollow)
			db.Set(addrFollow, msg)
			addrFollows = append(addrFollows, addrFollow)
		}
	}
	return addrFollows, nil
}

func ListenAddrHeightRoomFollows(ctx context.Context, addrs [][25]byte) (chan *AddrHeightRoomFollow, error) {
	if len(addrs) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, addr := range addrs {
		shard := client.GetByteShard32(addr[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addr[:])
	}
	shardConfigs := config.GetQueueShards()
	var addrRoomFollowChan = make(chan *AddrHeightRoomFollow)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(addrRoomFollowChan)
	})
	for shard, prefixes := range shardPrefixes {
		dbClient := client.NewClient(config.GetShardConfig(shard, shardConfigs).GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoAddrHeightRoomFollow, prefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db memo addr room follow by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var addrRoomFollow = new(AddrHeightRoomFollow)
				db.Set(addrRoomFollow, *msg)
				addrRoomFollowChan <- addrRoomFollow
			}
			cancelCtx.Cancel()
		}()
	}
	return addrRoomFollowChan, nil
}
