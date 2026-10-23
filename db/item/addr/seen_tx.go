package addr

import (
	"context"
	"fmt"
	"time"

	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
)

// SeenTx records that an address was party to a tx: an output paid to it, or
// an input spent from it. The key is the address's key bytes (25 for P2PKH
// and P2SH20, 33 for P2SH32, see wallet.LockAddr.Bytes), the seen time and
// the tx hash, so a prefix scan on the address bytes walks its history in
// time order. Inputs are always recorded under the 20-byte form (a P2SH32
// spend lands under the P2SH20 alias of its redeem script); P2shLink maps
// between the two.
type SeenTx struct {
	Addr   wallet.LockAddr
	Seen   time.Time
	TxHash [32]byte
}

const seenTxUidSuffixLength = 8 + 32

func (i *SeenTx) GetTopic() string {
	return db.TopicAddrSeenTx
}

func (i *SeenTx) GetShardSource() uint {
	return client.GenShardSource(i.Addr.Bytes())
}

func (i *SeenTx) GetUid() []byte {
	return jutil.CombineBytes(
		i.Addr.Bytes(),
		jutil.GetTimeByteNanoBig(i.Seen),
		jutil.ByteReverse(i.TxHash[:]),
	)
}

func (i *SeenTx) SetUid(uid []byte) {
	if len(uid) == 0 {
		return
	}
	// The address's version byte says how long it is; the rest is fixed.
	addrLen := wallet.LockAddrLength(uid[0])
	if len(uid) != addrLen+seenTxUidSuffixLength {
		return
	}
	addr, err := wallet.GetLockAddrFromBytes(uid[:addrLen])
	if err != nil {
		return
	}
	i.Addr = addr
	i.Seen = jutil.GetByteTimeNanoBig(uid[addrLen : addrLen+8])
	copy(i.TxHash[:], jutil.ByteReverse(uid[addrLen+8:]))
}

func (i *SeenTx) Serialize() []byte {
	return nil
}

func (i *SeenTx) Deserialize([]byte) {}

func GetSeenTxs(ctx context.Context, addr wallet.LockAddr, start []byte, limit uint32) ([]*SeenTx, error) {
	if limit == 0 || limit > uint32(client.ExLargeLimit) {
		limit = uint32(client.ExLargeLimit)
	}
	shardConfig := config.GetShardConfig(client.GenShardSource32(addr.Bytes()), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetByPrefix(ctx, db.TopicAddrSeenTx, client.Prefix{
		Prefix: addr.Bytes(),
		Start:  start,
		Limit:  limit,
	}); err != nil {
		return nil, fmt.Errorf("error getting db addr seen txs by prefix; %w", err)
	}
	var seenTxs = make([]*SeenTx, len(dbClient.Messages))
	for i := range dbClient.Messages {
		seenTxs[i] = new(SeenTx)
		db.Set(seenTxs[i], dbClient.Messages[i])
	}
	return seenTxs, nil
}

func ListenAddrSeenTxs(ctx context.Context, addrs []wallet.LockAddr) (chan *SeenTx, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range addrs {
		addrBytes := addrs[i].Bytes()
		shard := db.GetShardIdFromByte32(addrBytes)
		shardPrefixes[shard] = append(shardPrefixes[shard], addrBytes)
	}
	chanMessages, err := db.ListenPrefixes(ctx, db.TopicAddrSeenTx, shardPrefixes)
	if err != nil {
		return nil, fmt.Errorf("error getting listen prefixes for address seen tx; %w", err)
	}
	var chanSeenTxs = make(chan *SeenTx)
	go func() {
		defer close(chanSeenTxs)
		for msg := range chanMessages {
			var addrSeenTx = new(SeenTx)
			db.Set(addrSeenTx, *msg)
			select {
			case <-ctx.Done():
				return
			case chanSeenTxs <- addrSeenTx:
			}
		}
	}()
	return chanSeenTxs, nil
}
