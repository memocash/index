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

type AddrProfilePic struct {
	Addr   [25]byte
	Seen   time.Time
	TxHash [32]byte
	Pic    string
}

func (p *AddrProfilePic) GetTopic() string {
	return db.TopicMemoAddrProfilePic
}

func (p *AddrProfilePic) GetShardSource() uint {
	return client.GenShardSource(p.Addr[:])
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

func GetAddrProfilePics(ctx context.Context, addrs [][25]byte) ([]*AddrProfilePic, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range addrs {
		shard := db.GetShardIdFromByte32(addrs[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addrs[i][:])
	}
	shardConfigs := config.GetQueueShards()
	var addrProfilePics []*AddrProfilePic
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicMemoAddrProfilePic,
			Prefixes: prefixes,
			Max:      1,
			Newest:   true,
			Context:  ctx,
		}); err != nil {
			return nil, fmt.Errorf("error getting db addr memo profile pics by prefix; %w", err)
		}
		for _, msg := range dbClient.Messages {
			var addrProfilePic = new(AddrProfilePic)
			db.Set(addrProfilePic, msg)
			addrProfilePics = append(addrProfilePics, addrProfilePic)
		}
	}
	return addrProfilePics, nil
}

func ListenAddrProfilePics(ctx context.Context, addrs [][25]byte) (chan *AddrProfilePic, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range addrs {
		shard := db.GetShardIdFromByte32(addrs[i][:])
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
