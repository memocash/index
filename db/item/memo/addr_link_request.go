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

type AddrLinkRequest struct {
	Addr   [25]byte
	Seen   time.Time
	TxHash [32]byte
}

func (r *AddrLinkRequest) GetTopic() string {
	return db.TopicMemoAddrLinkRequest
}

func (r *AddrLinkRequest) GetShardSource() uint {
	return client.GenShardSource(r.Addr[:])
}

func (r *AddrLinkRequest) GetUid() []byte {
	return jutil.CombineBytes(
		r.Addr[:],
		jutil.GetTimeByteNanoBig(r.Seen),
		jutil.ByteReverse(r.TxHash[:]),
	)
}

func (r *AddrLinkRequest) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(r.Addr[:], uid[:25])
	r.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(r.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (r *AddrLinkRequest) Serialize() []byte {
	return nil
}

func (r *AddrLinkRequest) Deserialize([]byte) {}

func GetAddrLinkRequests(ctx context.Context, addrs [][25]byte) ([]*AddrLinkRequest, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range addrs {
		shard := client.GenShardSource32(addrs[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addrs[i][:])
	}
	shardConfigs := config.GetQueueShards()
	var addrLinkRequests []*AddrLinkRequest
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicMemoAddrLinkRequest,
			Prefixes: prefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, fmt.Errorf("error getting db addr memo link requests by prefix; %w", err)
		}
		for _, msg := range dbClient.Messages {
			var addrLinkRequest = new(AddrLinkRequest)
			db.Set(addrLinkRequest, msg)
			addrLinkRequests = append(addrLinkRequests, addrLinkRequest)
		}
	}
	return addrLinkRequests, nil
}
