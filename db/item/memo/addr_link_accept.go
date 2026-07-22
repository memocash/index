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

type AddrLinkAccept struct {
	Addr          [25]byte
	Seen          time.Time
	TxHash        [32]byte
	RequestTxHash [32]byte
	Message       string
}

func (a *AddrLinkAccept) GetTopic() string {
	return db.TopicMemoAddrLinkAccept
}

func (a *AddrLinkAccept) GetShardSource() uint {
	return client.GenShardSource(a.Addr[:])
}

func (a *AddrLinkAccept) GetUid() []byte {
	return jutil.CombineBytes(
		a.Addr[:],
		jutil.GetTimeByteNanoBig(a.Seen),
		jutil.ByteReverse(a.TxHash[:]),
	)
}

func (a *AddrLinkAccept) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(a.Addr[:], uid[:25])
	a.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(a.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (a *AddrLinkAccept) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(a.RequestTxHash[:]),
		[]byte(a.Message),
	)
}

func (a *AddrLinkAccept) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength {
		return
	}
	copy(a.RequestTxHash[:], jutil.ByteReverse(data[:memo.TxHashLength]))
	a.Message = string(data[memo.TxHashLength:])
}

func GetAddrLinkAccepts(ctx context.Context, addrs [][25]byte) ([]*AddrLinkAccept, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicMemoAddrLinkAccept, db.ShardPrefixesAddrs(addrs),
		client.NewOptionPrefixLimit(client.ExLargeLimit))
	if err != nil {
		return nil, fmt.Errorf("error getting db addr memo link accepts by prefixes; %w", err)
	}
	var addrLinkAccepts = make([]*AddrLinkAccept, len(messages))
	for i := range messages {
		addrLinkAccepts[i] = new(AddrLinkAccept)
		db.Set(addrLinkAccepts[i], messages[i])
	}
	return addrLinkAccepts, nil
}
