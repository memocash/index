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

type AddrLinkRevoke struct {
	Addr         [25]byte
	Seen         time.Time
	TxHash       [32]byte
	AcceptTxHash [32]byte
	Message      string
}

func (r *AddrLinkRevoke) GetTopic() string {
	return db.TopicMemoAddrLinkRevoke
}

func (r *AddrLinkRevoke) GetShardSource() uint {
	return client.GenShardSource(r.Addr[:])
}

func (r *AddrLinkRevoke) GetUid() []byte {
	return jutil.CombineBytes(
		r.Addr[:],
		jutil.GetTimeByteNanoBig(r.Seen),
		jutil.ByteReverse(r.TxHash[:]),
	)
}

func (r *AddrLinkRevoke) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(r.Addr[:], uid[:25])
	r.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(r.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (r *AddrLinkRevoke) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(r.AcceptTxHash[:]),
		[]byte(r.Message),
	)
}

func (r *AddrLinkRevoke) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength {
		return
	}
	copy(r.AcceptTxHash[:], jutil.ByteReverse(data[:memo.TxHashLength]))
	r.Message = string(data[memo.TxHashLength:])
}

func GetAddrLinkRevokes(ctx context.Context, addrs [][25]byte) ([]*AddrLinkRevoke, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicMemoAddrLinkRevoke, db.ShardPrefixesAddrs(addrs),
		client.NewOptionPrefixLimit(client.ExLargeLimit))
	if err != nil {
		return nil, fmt.Errorf("error getting db addr memo link revokes by prefixes; %w", err)
	}
	var addrLinkRevokes = make([]*AddrLinkRevoke, len(messages))
	for i := range messages {
		addrLinkRevokes[i] = new(AddrLinkRevoke)
		db.Set(addrLinkRevokes[i], messages[i])
	}
	return addrLinkRevokes, nil
}
