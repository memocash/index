package addr

import (
	"context"
	"fmt"

	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// P2shLink maps one encoding of a P2SH redeem script to the other: the P2SH20
// address (HASH160) to the P2SH32 address (HASH256) and back, one row per
// direction. A P2SH20 spend and a P2SH32 spend have identical unlock scripts,
// so the saver cannot tell which the output used without reading it; instead
// every P2SH-shaped spend is recorded under the P2SH20 alias and the link is
// written, so a P2SH32 address's spends are found through its alias. The
// link is written when the redeem script is first seen in an unlock script,
// which is also the earliest any spend can exist: no link means no spends
// yet. Either encoding resolves to the same history, in both directions.
type P2shLink struct {
	Addr  wallet.LockAddr
	Other wallet.LockAddr
}

func (l *P2shLink) GetTopic() string {
	return db.TopicAddrP2shLink
}

func (l *P2shLink) GetShardSource() uint {
	return client.GenShardSource(l.Addr.Bytes())
}

func (l *P2shLink) GetUid() []byte {
	return l.Addr.Bytes()
}

func (l *P2shLink) SetUid(uid []byte) {
	addr, err := wallet.GetLockAddrFromBytes(uid)
	if err != nil {
		return
	}
	l.Addr = addr
}

func (l *P2shLink) Serialize() []byte {
	return l.Other.Bytes()
}

func (l *P2shLink) Deserialize(data []byte) {
	other, err := wallet.GetLockAddrFromBytes(data)
	if err != nil {
		return
	}
	l.Other = other
}

// NewP2shLinks returns the two rows linking a redeem script's P2SH20 and
// P2SH32 addresses.
func NewP2shLinks(addr wallet.Addr, addr32 wallet.Addr32) []db.Object {
	return []db.Object{
		&P2shLink{Addr: addr.LockAddr(), Other: addr32.LockAddr()},
		&P2shLink{Addr: addr32.LockAddr(), Other: addr.LockAddr()},
	}
}

// GetP2shLinks returns the links for the given addresses that have one.
func GetP2shLinks(ctx context.Context, addrs []wallet.LockAddr) ([]*P2shLink, error) {
	var shardUids = make(map[uint32][][]byte)
	for _, addr := range addrs {
		addrBytes := addr.Bytes()
		shard := db.GetShardIdFromByte32(addrBytes)
		shardUids[shard] = append(shardUids[shard], addrBytes)
	}
	messages, err := db.GetSpecific(ctx, db.TopicAddrP2shLink, shardUids)
	if err != nil {
		return nil, fmt.Errorf("error getting addr p2sh links; %w", err)
	}
	var links = make([]*P2shLink, len(messages))
	for i := range messages {
		links[i] = new(P2shLink)
		db.Set(links[i], messages[i])
	}
	return links, nil
}

// ListenP2shLinks delivers links as they are written for the given addresses,
// so a subscriber can pick up an alias that appears after it started.
func ListenP2shLinks(ctx context.Context, addrs []wallet.LockAddr) (chan *P2shLink, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, addr := range addrs {
		addrBytes := addr.Bytes()
		shard := db.GetShardIdFromByte32(addrBytes)
		shardPrefixes[shard] = append(shardPrefixes[shard], addrBytes)
	}
	chanMessages, err := db.ListenPrefixes(ctx, db.TopicAddrP2shLink, shardPrefixes)
	if err != nil {
		return nil, fmt.Errorf("error getting listen prefixes for addr p2sh links; %w", err)
	}
	var chanLinks = make(chan *P2shLink)
	go func() {
		defer close(chanLinks)
		for msg := range chanMessages {
			var link = new(P2shLink)
			db.Set(link, *msg)
			select {
			case <-ctx.Done():
				return
			case chanLinks <- link:
			}
		}
	}()
	return chanLinks, nil
}
