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

type AddrLinkRequestParent struct {
	ParentAddr [25]byte
	Seen       time.Time
	TxHash     [32]byte
	Addr       [25]byte
	Message    string
}

func (r *AddrLinkRequestParent) GetTopic() string {
	return db.TopicMemoAddrLinkRequestParent
}

func (r *AddrLinkRequestParent) GetShardSource() uint {
	return client.GenShardSource(r.ParentAddr[:])
}

func (r *AddrLinkRequestParent) GetUid() []byte {
	return jutil.CombineBytes(
		r.ParentAddr[:],
		jutil.GetTimeByteNanoBig(r.Seen),
		jutil.ByteReverse(r.TxHash[:]),
	)
}

func (r *AddrLinkRequestParent) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(r.ParentAddr[:], uid[:25])
	r.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(r.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (r *AddrLinkRequestParent) Serialize() []byte {
	return jutil.CombineBytes(
		r.Addr[:],
		[]byte(r.Message),
	)
}

func (r *AddrLinkRequestParent) Deserialize(data []byte) {
	if len(data) < memo.AddressLength {
		return
	}
	copy(r.Addr[:], data[:memo.AddressLength])
	r.Message = string(data[memo.AddressLength:])
}

func GetAddrLinkRequestParents(ctx context.Context, parentAddrs [][25]byte) ([]*AddrLinkRequestParent, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicMemoAddrLinkRequestParent, db.ShardPrefixesAddrs(parentAddrs))
	if err != nil {
		return nil, fmt.Errorf("error getting db addr memo link request parents by prefixes; %w", err)
	}
	var addrLinkRequestParents = make([]*AddrLinkRequestParent, len(messages))
	for i := range messages {
		addrLinkRequestParents[i] = new(AddrLinkRequestParent)
		db.Set(addrLinkRequestParents[i], messages[i])
	}
	return addrLinkRequestParents, nil
}

func GetAddrLinkRequestParentsSingle(ctx context.Context, parentAddr [25]byte, start time.Time) ([]*AddrLinkRequestParent, error) {
	dbClient := db.GetShardClient(client.GenShardSource32(parentAddr[:]))
	prefix := client.NewPrefix(parentAddr[:])
	if !jutil.IsTimeZero(start) {
		prefix.Start = jutil.CombineBytes(parentAddr[:], jutil.GetTimeByteNanoBig(start))
	}
	if err := dbClient.GetByPrefix(ctx, db.TopicMemoAddrLinkRequestParent, prefix, client.OptionExLargeLimit()); err != nil {
		return nil, fmt.Errorf("error getting db addr memo link request parent by prefix; %w", err)
	}
	var addrLinkRequestParents = make([]*AddrLinkRequestParent, len(dbClient.Messages))
	for i := range dbClient.Messages {
		addrLinkRequestParents[i] = new(AddrLinkRequestParent)
		db.Set(addrLinkRequestParents[i], dbClient.Messages[i])
	}
	return addrLinkRequestParents, nil
}
