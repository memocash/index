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

type AddrHeightProfile struct {
	Addr    [25]byte
	Height  int64
	TxHash  [32]byte
	Profile string
}

func (p *AddrHeightProfile) GetTopic() string {
	return db.TopicMemoAddrHeightProfile
}

func (p *AddrHeightProfile) GetShard() uint {
	return client.GetByteShard(p.Addr[:])
}

func (p *AddrHeightProfile) GetUid() []byte {
	return jutil.CombineBytes(
		p.Addr[:],
		jutil.ByteFlip(jutil.GetInt64DataBig(p.Height)),
		jutil.ByteReverse(p.TxHash[:]),
	)
}

func (p *AddrHeightProfile) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(p.Addr[:], uid[:25])
	p.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[25:33]))
	copy(p.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (p *AddrHeightProfile) Serialize() []byte {
	return []byte(p.Profile)
}

func (p *AddrHeightProfile) Deserialize(data []byte) {
	p.Profile = string(data)
}

func GetAddrHeightProfile(ctx context.Context, addr [25]byte) (*AddrHeightProfile, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(addr[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoAddrHeightProfile,
		Prefixes: [][]byte{addr[:]},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db addr memo profile by prefix", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, jerr.Get("error no addr memo profiles found", client.EntryNotFoundError)
	}
	var addrProfile = new(AddrHeightProfile)
	db.Set(addrProfile, dbClient.Messages[0])
	return addrProfile, nil
}

func ListenAddrHeightProfiles(ctx context.Context, addrs [][25]byte) (chan *AddrHeightProfile, error) {
	if len(addrs) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, addr := range addrs {
		shard := client.GetByteShard32(addr[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addr[:])
	}
	shardConfigs := config.GetQueueShards()
	var addrProfileChan = make(chan *AddrHeightProfile)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(addrProfileChan)
	})
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoAddrHeightProfile, prefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db addr memo profile by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var addrProfile = new(AddrHeightProfile)
				db.Set(addrProfile, *msg)
				addrProfileChan <- addrProfile
			}
			cancelCtx.Cancel()
		}()
	}
	return addrProfileChan, nil
}
