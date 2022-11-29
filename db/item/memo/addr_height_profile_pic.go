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

type AddrHeightProfilePic struct {
	Addr   [25]byte
	Height int64
	TxHash [32]byte
	Pic    string
}

func (p *AddrHeightProfilePic) GetTopic() string {
	return db.TopicMemoAddrHeightProfilePic
}

func (p *AddrHeightProfilePic) GetShard() uint {
	return client.GetByteShard(p.Addr[:])
}

func (p *AddrHeightProfilePic) GetUid() []byte {
	return jutil.CombineBytes(
		p.Addr[:],
		jutil.ByteFlip(jutil.GetInt64DataBig(p.Height)),
		jutil.ByteReverse(p.TxHash[:]),
	)
}

func (p *AddrHeightProfilePic) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(p.Addr[:], uid[:25])
	p.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[25:33]))
	copy(p.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (p *AddrHeightProfilePic) Serialize() []byte {
	return []byte(p.Pic)
}

func (p *AddrHeightProfilePic) Deserialize(data []byte) {
	p.Pic = string(data)
}

func GetAddrHeightProfilePic(ctx context.Context, addr [25]byte) (*AddrHeightProfilePic, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(addr[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoAddrHeightProfilePic,
		Prefixes: [][]byte{addr[:]},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db addr memo profile pic by prefix", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, jerr.Get("error no addr memo profile pics found", client.EntryNotFoundError)
	}
	var addrProfilePic = new(AddrHeightProfilePic)
	db.Set(addrProfilePic, dbClient.Messages[0])
	return addrProfilePic, nil
}

func ListenAddrHeightProfilePics(ctx context.Context, addrs [][25]byte) (chan *AddrHeightProfilePic, error) {
	if len(addrs) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, addr := range addrs {
		shard := client.GetByteShard32(addr[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addr[:])
	}
	shardConfigs := config.GetQueueShards()
	var addrProfilePicChan = make(chan *AddrHeightProfilePic)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(addrProfilePicChan)
	})
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoAddrHeightProfilePic, prefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db addr memo profile pic by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var addrProfilePic = new(AddrHeightProfilePic)
				db.Set(addrProfilePic, *msg)
				addrProfilePicChan <- addrProfilePic
			}
			cancelCtx.Cancel()
		}()
	}
	return addrProfilePicChan, nil
}
