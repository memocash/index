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

type AddrPost struct {
	Addr   [25]byte
	Seen   time.Time
	TxHash [32]byte
}

func (p *AddrPost) GetTopic() string {
	return db.TopicMemoAddrPost
}

func (p *AddrPost) GetShardSource() uint {
	return client.GenShardSource(p.Addr[:])
}

func (p *AddrPost) GetUid() []byte {
	return jutil.CombineBytes(
		p.Addr[:],
		jutil.GetTimeByteNanoBig(p.Seen),
		jutil.ByteReverse(p.TxHash[:]),
	)
}

func (p *AddrPost) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(p.Addr[:], uid[:25])
	p.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(p.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (p *AddrPost) Serialize() []byte {
	return nil
}

func (p *AddrPost) Deserialize([]byte) {}

func GetSingleAddrPosts(ctx context.Context, addr [25]byte, newest bool, start time.Time) ([]*AddrPost, error) {
	dbClient := db.GetShardClient(client.GenShardSource32(addr[:]))
	prefix := client.NewPrefix(addr[:])
	if !jutil.IsTimeZero(start) {
		prefix.Start = jutil.CombineBytes(addr[:], jutil.GetTimeByteNanoBig(start))
	}
	var opts = []client.Option{client.OptionExLargeLimit(), client.NewOptionOrder(newest)}
	err := dbClient.GetByPrefix(ctx, db.TopicMemoAddrPost, prefix, opts...)
	if err != nil {
		return nil, fmt.Errorf("error getting db addr memo post by prefix; %w", err)
	}
	var addrPosts = make([]*AddrPost, len(dbClient.Messages))
	for i := range dbClient.Messages {
		addrPosts[i] = new(AddrPost)
		db.Set(addrPosts[i], dbClient.Messages[i])
	}
	return addrPosts, nil
}

func GetAddrPosts(ctx context.Context, addrs [][25]byte, newest bool) ([]*AddrPost, error) {
	shardPrefixes := db.ShardPrefixesAddrs(addrs)
	var opts = []client.Option{client.OptionExLargeLimit(), client.NewOptionOrder(newest)}
	messages, err := db.GetByPrefixes(ctx, db.TopicMemoAddrPost, shardPrefixes, opts...)
	if err != nil {
		return nil, fmt.Errorf("error getting db addr memo post by prefix; %w", err)
	}
	var addrPosts = make([]*AddrPost, len(messages))
	for i := range messages {
		addrPosts[i] = new(AddrPost)
		db.Set(addrPosts[i], messages[i])
	}
	return addrPosts, nil
}
