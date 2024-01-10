package memo

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
	"time"
)

type AddrProfilePic struct {
	Addr   [25]byte
	Seen   time.Time
	TxHash [32]byte
	Pic    string
}

func (p *AddrProfilePic) GetTopic() string {
	return db.TopicMemoAddrProfilePic
}

func (p *AddrProfilePic) GetShard() uint {
	return client.GetByteShard(p.Addr[:])
}

func (p *AddrProfilePic) GetUid() []byte {
	return jutil.CombineBytes(
		p.Addr[:],
		jutil.GetTimeByteNanoBig(p.Seen),
		jutil.ByteReverse(p.TxHash[:]),
	)
}

func (p *AddrProfilePic) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(p.Addr[:], uid[:25])
	p.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(p.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (p *AddrProfilePic) Serialize() []byte {
	return []byte(p.Pic)
}

func (p *AddrProfilePic) Deserialize(data []byte) {
	p.Pic = string(data)
}

func GetAddrProfilePic(ctx context.Context, addr [25]byte) (*AddrProfilePic, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(addr[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoAddrProfilePic,
		Prefixes: [][]byte{addr[:]},
		Max:      1,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db addr memo profile pic by prefix", err)
	}
	if len(dbClient.Messages) == 0 {
		return nil, jerr.Get("error no addr memo profile pics found", client.EntryNotFoundError)
	}
	var addrProfilePic = new(AddrProfilePic)
	db.Set(addrProfilePic, dbClient.Messages[0])
	return addrProfilePic, nil
}

func ListenAddrProfilePics(ctx context.Context, addrs [][25]byte) (chan *AddrProfilePic, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range addrs {
		shard := db.GetShardByte32(addrs[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addrs[i][:])
	}
	chanMessages, err := db.ListenPrefixes(ctx, db.TopicMemoAddrProfilePic, shardPrefixes)
	if err != nil {
		return nil, fmt.Errorf("error getting listen prefixes for memo addr profile pics; %w", err)
	}
	var addrProfilePicChan = make(chan *AddrProfilePic)
	go func() {
		for {
			msg, ok := <-chanMessages
			if !ok {
				return
			}
			var addrProfilePic = new(AddrProfilePic)
			db.Set(addrProfilePic, *msg)
			addrProfilePicChan <- addrProfilePic
		}
	}()
	return addrProfilePicChan, nil
}
