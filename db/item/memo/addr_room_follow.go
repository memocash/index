package memo

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
	"time"
)

type AddrRoomFollow struct {
	Addr     [25]byte
	Seen     time.Time
	TxHash   [32]byte
	Unfollow bool
	Room     string
}

func (f *AddrRoomFollow) GetTopic() string {
	return db.TopicMemoAddrRoomFollow
}

func (f *AddrRoomFollow) GetShardSource() uint {
	return client.GenShardSource(f.Addr[:])
}

func (f *AddrRoomFollow) GetUid() []byte {
	return jutil.CombineBytes(
		f.Addr[:],
		jutil.GetTimeByteNanoBig(f.Seen),
		jutil.ByteReverse(f.TxHash[:]),
	)
}

func (f *AddrRoomFollow) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(f.Addr[:], uid[:25])
	f.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(f.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (f *AddrRoomFollow) Serialize() []byte {
	var unfollow byte
	if f.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		[]byte(f.Room),
	)
}

func (f *AddrRoomFollow) Deserialize(data []byte) {
	if len(data) < 1 {
		return
	}
	f.Unfollow = data[0] == 1
	f.Room = string(data[1:])
}

func GetAddrRoomFollows(ctx context.Context, addrs [][25]byte) ([]*AddrRoomFollow, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicMemoAddrRoomFollow, db.ShardPrefixesAddrs(addrs))
	if err != nil {
		return nil, fmt.Errorf("error getting db memo addr room follow by prefix; %w", err)
	}
	var addrFollows = make([]*AddrRoomFollow, len(messages))
	for i := range messages {
		addrFollows[i] = new(AddrRoomFollow)
		db.Set(addrFollows[i], messages[i])
	}
	return addrFollows, nil
}

func ListenAddrRoomFollows(ctx context.Context, addrs [][25]byte) (chan *AddrRoomFollow, error) {
	if len(addrs) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, addr := range addrs {
		shard := client.GenShardSource32(addr[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addr[:])
	}
	shardConfigs := config.GetQueueShards()
	var addrRoomFollowChan = make(chan *AddrRoomFollow)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(addrRoomFollowChan)
	})
	for shard, prefixes := range shardPrefixes {
		dbClient := client.NewClient(config.GetShardConfig(shard, shardConfigs).GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoAddrRoomFollow, prefixes)
		if err != nil {
			return nil, fmt.Errorf("error listening to db memo addr room follow by prefix; %w", err)
		}
		go func() {
			for msg := range chanMessage {
				var addrRoomFollow = new(AddrRoomFollow)
				db.Set(addrRoomFollow, *msg)
				addrRoomFollowChan <- addrRoomFollow
			}
			cancelCtx.Cancel()
		}()
	}
	return addrRoomFollowChan, nil
}
