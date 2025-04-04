package memo

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"time"
)

type AddrFollow struct {
	Addr       [25]byte
	Seen       time.Time
	TxHash     [32]byte
	Unfollow   bool
	FollowAddr [25]byte
}

func (f *AddrFollow) GetTopic() string {
	return db.TopicMemoAddrFollow
}

func (f *AddrFollow) GetShardSource() uint {
	return client.GenShardSource(f.Addr[:])
}

func (f *AddrFollow) GetUid() []byte {
	return jutil.CombineBytes(
		f.Addr[:],
		jutil.GetTimeByteNanoBig(f.Seen),
		jutil.ByteReverse(f.TxHash[:]),
	)
}

func (f *AddrFollow) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(f.Addr[:], uid[:25])
	f.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(f.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (f *AddrFollow) Serialize() []byte {
	var unfollow byte
	if f.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		f.FollowAddr[:],
	)
}

func (f *AddrFollow) Deserialize(data []byte) {
	if len(data) < 1+memo.AddressLength {
		return
	}
	f.Unfollow = data[0] == 1
	copy(f.FollowAddr[:], data[1:26])
}

func GetAddrFollows(ctx context.Context, addrs [][25]byte) ([]*AddrFollow, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicMemoAddrFollow, db.ShardPrefixesAddrs(addrs))
	if err != nil {
		return nil, fmt.Errorf("error getting db addr memo follow by prefixes; %w", err)
	}
	var addrFollows = make([]*AddrFollow, len(messages))
	for i := range messages {
		addrFollows[i] = new(AddrFollow)
		db.Set(addrFollows[i], messages[i])
	}
	return addrFollows, nil
}

func GetAddrFollowsSingle(ctx context.Context, addr [25]byte, start time.Time) ([]*AddrFollow, error) {
	dbClient := db.GetShardClient(client.GenShardSource32(addr[:]))
	prefix := client.NewPrefix(addr[:])
	if !jutil.IsTimeZero(start) {
		prefix.Start = jutil.CombineBytes(addr[:], jutil.GetTimeByteNanoBig(start))
	}
	if err := dbClient.GetByPrefix(ctx, db.TopicMemoAddrFollow, prefix, client.OptionExLargeLimit()); err != nil {
		return nil, fmt.Errorf("error getting db addr memo follow by prefix; %w", err)
	}
	var addrFollows = make([]*AddrFollow, len(dbClient.Messages))
	for i := range dbClient.Messages {
		addrFollows[i] = new(AddrFollow)
		db.Set(addrFollows[i], dbClient.Messages[i])
	}
	return addrFollows, nil
}

func ListenAddrFollows(ctx context.Context, addrs [][25]byte) (chan *AddrFollow, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range addrs {
		shard := db.GetShardIdFromByte32(addrs[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addrs[i][:])
	}
	chanMessages, err := db.ListenPrefixes(ctx, db.TopicMemoAddrFollow, shardPrefixes)
	if err != nil {
		return nil, fmt.Errorf("error getting listen prefixes for memo addr follows; %w", err)
	}
	var addrFollowChan = make(chan *AddrFollow)
	go func() {
		for {
			msg, ok := <-chanMessages
			if !ok {
				return
			}
			var addrProfile = new(AddrFollow)
			db.Set(addrProfile, *msg)
			addrFollowChan <- addrProfile
		}
	}()
	return addrFollowChan, nil
}
