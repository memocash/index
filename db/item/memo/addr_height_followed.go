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

type AddrHeightFollowed struct {
	FollowAddr [25]byte
	Height     int64
	TxHash     [32]byte
	Addr       [25]byte
	Unfollow   bool
}

func (f *AddrHeightFollowed) GetTopic() string {
	return db.TopicMemoAddrHeightFollowed
}

func (f *AddrHeightFollowed) GetShard() uint {
	return client.GetByteShard(f.FollowAddr[:])
}

func (f *AddrHeightFollowed) GetUid() []byte {
	return jutil.CombineBytes(
		f.FollowAddr[:],
		jutil.ByteFlip(jutil.GetInt64DataBig(f.Height)),
		jutil.ByteReverse(f.TxHash[:]),
	)
}

func (f *AddrHeightFollowed) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(f.FollowAddr[:], uid[:25])
	f.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[25:33]))
	copy(f.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (f *AddrHeightFollowed) Serialize() []byte {
	var unfollow byte
	if f.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		f.Addr[:],
	)
}

func (f *AddrHeightFollowed) Deserialize(data []byte) {
	if len(data) < 1+memo.AddressLength {
		return
	}
	f.Unfollow = data[0] == 1
	copy(f.Addr[:], data[1:26])
}

func GetAddrHeightFolloweds(ctx context.Context, followAddresses [][25]byte) ([]*AddrHeightFollowed, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, followAddress := range followAddresses {
		shard := client.GetByteShard32(followAddress[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], followAddress[:])
	}
	shardConfigs := config.GetQueueShards()
	var addrFolloweds []*AddrHeightFollowed
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicMemoAddrHeightFollowed,
			Prefixes: prefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db addr memo followed by prefix", err)
		}
		for _, msg := range dbClient.Messages {
			var addrFollowed = new(AddrHeightFollowed)
			db.Set(addrFollowed, msg)
			addrFolloweds = append(addrFolloweds, addrFollowed)
		}
	}
	return addrFolloweds, nil
}

func GetAddrHeightFollowedsSingle(ctx context.Context, followAddr [25]byte, start int64) ([]*AddrHeightFollowed, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(followAddr[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var startByte []byte
	if start != 0 {
		startByte = jutil.CombineBytes(followAddr[:], jutil.ByteFlip(jutil.GetInt64DataBig(start)))
	} else {
		startByte = followAddr[:]
	}
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoAddrHeightFollowed,
		Prefixes: [][]byte{followAddr[:]},
		Start:    startByte,
		Max:      client.ExLargeLimit,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db addr memo follow by prefix", err)
	}
	var addrFolloweds = make([]*AddrHeightFollowed, len(dbClient.Messages))
	for i := range dbClient.Messages {
		addrFolloweds[i] = new(AddrHeightFollowed)
		db.Set(addrFolloweds[i], dbClient.Messages[i])
	}
	return addrFolloweds, nil
}

func ListenAddrHeightFolloweds(ctx context.Context, followAddrs [][25]byte) (chan *AddrHeightFollowed, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, followAddr := range followAddrs {
		shard := client.GetByteShard32(followAddr[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], followAddr[:])
	}
	shardConfigs := config.GetQueueShards()
	var addrFollowedChan = make(chan *AddrHeightFollowed)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(addrFollowedChan)
	})
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		chanMessage, err := client.NewClient(shardConfig.GetHost()).
			Listen(cancelCtx.Context, db.TopicMemoAddrHeightFollowed, prefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db addr memo followeds by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var addrFollowed = new(AddrHeightFollowed)
				db.Set(addrFollowed, *msg)
				addrFollowedChan <- addrFollowed
			}
			cancelCtx.Cancel()
		}()
	}
	return addrFollowedChan, nil
}
