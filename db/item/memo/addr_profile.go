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

type AddrProfile struct {
	Addr    [25]byte
	Seen    time.Time
	TxHash  [32]byte
	Profile string
}

func (p *AddrProfile) GetTopic() string {
	return db.TopicMemoAddrProfile
}

func (p *AddrProfile) GetShard() uint {
	return client.GetByteShard(p.Addr[:])
}

func (p *AddrProfile) GetUid() []byte {
	return jutil.CombineBytes(
		p.Addr[:],
		jutil.GetTimeByteBig(p.Seen),
		jutil.ByteReverse(p.TxHash[:]),
	)
}

func (p *AddrProfile) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(p.Addr[:], uid[:25])
	p.Seen = jutil.GetByteTimeBig(uid[25:33])
	copy(p.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (p *AddrProfile) Serialize() []byte {
	return []byte(p.Profile)
}

func (p *AddrProfile) Deserialize(data []byte) {
	p.Profile = string(data)
}

func GetAddrProfile(ctx context.Context, addr [25]byte) (*AddrProfile, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(addr[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoAddrProfile,
		Prefixes: [][]byte{addr[:]},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db addr memo profile by prefix", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, jerr.Get("error no addr memo profiles found", client.EntryNotFoundError)
	}
	var addrProfile = new(AddrProfile)
	db.Set(addrProfile, dbClient.Messages[0])
	return addrProfile, nil
}

func ListenAddrProfiles(ctx context.Context, addrs [][25]byte) (chan *AddrProfile, error) {
	if len(addrs) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, addr := range addrs {
		shard := client.GetByteShard32(addr[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addr[:])
	}
	shardConfigs := config.GetQueueShards()
	var addrProfileChan = make(chan *AddrProfile)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(addrProfileChan)
	})
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoAddrProfile, prefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db addr memo profile by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var addrProfile = new(AddrProfile)
				db.Set(addrProfile, *msg)
				addrProfileChan <- addrProfile
			}
			cancelCtx.Cancel()
		}()
	}
	return addrProfileChan, nil
}
