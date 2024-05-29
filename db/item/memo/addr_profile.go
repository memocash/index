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

type AddrProfile struct {
	Addr    [25]byte
	Seen    time.Time
	TxHash  [32]byte
	Profile string
}

func (p *AddrProfile) GetTopic() string {
	return db.TopicMemoAddrProfile
}

func (p *AddrProfile) GetShardSource() uint {
	return client.GenShardSource(p.Addr[:])
}

func (p *AddrProfile) GetUid() []byte {
	return jutil.CombineBytes(
		p.Addr[:],
		jutil.GetTimeByteNanoBig(p.Seen),
		jutil.ByteReverse(p.TxHash[:]),
	)
}

func (p *AddrProfile) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(p.Addr[:], uid[:25])
	p.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(p.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (p *AddrProfile) Serialize() []byte {
	return []byte(p.Profile)
}

func (p *AddrProfile) Deserialize(data []byte) {
	p.Profile = string(data)
}

func GetAddrProfiles(ctx context.Context, addrs [][25]byte) ([]*AddrProfile, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range addrs {
		shard := db.GetShardIdFromByte32(addrs[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addrs[i][:])
	}
	shardConfigs := config.GetQueueShards()
	var addrProfiles []*AddrProfile
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicMemoAddrProfile,
			Prefixes: prefixes,
			Max:      1,
			Newest:   true,
			Context:  ctx,
		}); err != nil {
			return nil, fmt.Errorf("error getting db addr memo profiles by prefix; %w", err)
		}
		for _, msg := range dbClient.Messages {
			var addrProfile = new(AddrProfile)
			db.Set(addrProfile, msg)
			addrProfiles = append(addrProfiles, addrProfile)
		}
	}
	return addrProfiles, nil
}

func ListenAddrProfiles(ctx context.Context, addrs [][25]byte) (chan *AddrProfile, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range addrs {
		shard := db.GetShardIdFromByte32(addrs[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addrs[i][:])
	}
	chanMessages, err := db.ListenPrefixes(ctx, db.TopicMemoAddrProfile, shardPrefixes)
	if err != nil {
		return nil, fmt.Errorf("error getting listen prefixes for memo addr profiles; %w", err)
	}
	var addrProfileChan = make(chan *AddrProfile)
	go func() {
		for {
			msg, ok := <-chanMessages
			if !ok {
				return
			}
			var addrProfile = new(AddrProfile)
			db.Set(addrProfile, *msg)
			addrProfileChan <- addrProfile
		}
	}()
	return addrProfileChan, nil
}
