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

type AddrFollowed struct {
	FollowAddr [25]byte
	Seen       time.Time
	TxHash     [32]byte
	Addr       [25]byte
	Unfollow   bool
}

func (f *AddrFollowed) GetTopic() string {
	return db.TopicMemoAddrFollowed
}

func (f *AddrFollowed) GetShard() uint {
	return client.GetByteShard(f.FollowAddr[:])
}

func (f *AddrFollowed) GetUid() []byte {
	return jutil.CombineBytes(
		f.FollowAddr[:],
		jutil.GetTimeByteNanoBig(f.Seen),
		jutil.ByteReverse(f.TxHash[:]),
	)
}

func (f *AddrFollowed) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(f.FollowAddr[:], uid[:25])
	f.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(f.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (f *AddrFollowed) Serialize() []byte {
	var unfollow byte
	if f.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		f.Addr[:],
	)
}

func (f *AddrFollowed) Deserialize(data []byte) {
	if len(data) < 1+memo.AddressLength {
		return
	}
	f.Unfollow = data[0] == 1
	copy(f.Addr[:], data[1:26])
}

func GetAddrFolloweds(ctx context.Context, followAddresses [][25]byte) ([]*AddrFollowed, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range followAddresses {
		shard := client.GetByteShard32(followAddresses[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], followAddresses[i][:])
	}
	shardConfigs := config.GetQueueShards()
	var addrFolloweds []*AddrFollowed
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicMemoAddrFollowed,
			Prefixes: prefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db addr memo followed by prefix", err)
		}
		for _, msg := range dbClient.Messages {
			var addrFollowed = new(AddrFollowed)
			db.Set(addrFollowed, msg)
			addrFolloweds = append(addrFolloweds, addrFollowed)
		}
	}
	return addrFolloweds, nil
}

func GetAddrFollowedsSingle(ctx context.Context, followAddr [25]byte, start time.Time) ([]*AddrFollowed, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(followAddr[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var startByte []byte
	if !jutil.IsTimeZero(start) {
		startByte = jutil.CombineBytes(followAddr[:], jutil.GetTimeByteNanoBig(start))
	} else {
		startByte = followAddr[:]
	}
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoAddrFollowed,
		Prefixes: [][]byte{followAddr[:]},
		Start:    startByte,
		Max:      client.ExLargeLimit,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db addr memo follow by prefix", err)
	}
	var addrFolloweds = make([]*AddrFollowed, len(dbClient.Messages))
	for i := range dbClient.Messages {
		addrFolloweds[i] = new(AddrFollowed)
		db.Set(addrFolloweds[i], dbClient.Messages[i])
	}
	return addrFolloweds, nil
}

func ListenAddrFolloweds(ctx context.Context, followAddrs [][25]byte) (chan *AddrFollowed, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range followAddrs {
		shard := client.GetByteShard32(followAddrs[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], followAddrs[i][:])
	}
	shardConfigs := config.GetQueueShards()
	var addrFollowedChan = make(chan *AddrFollowed)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(addrFollowedChan)
	})
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		chanMessage, err := client.NewClient(shardConfig.GetHost()).
			Listen(cancelCtx.Context, db.TopicMemoAddrFollowed, prefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db addr memo followeds by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var addrFollowed = new(AddrFollowed)
				db.Set(addrFollowed, *msg)
				addrFollowedChan <- addrFollowed
			}
			cancelCtx.Cancel()
		}()
	}
	return addrFollowedChan, nil
}
