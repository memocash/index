package sub

import (
	"context"
	"fmt"
	"log"

	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/graph/attach"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// ListenAddrSeenTxs streams the seen-tx rows for a set of addresses, each
// delivered once per tx. A P2SH address's history spans two keys (its own
// encoding for outputs, the P2SH20 alias of its redeem script for spends),
// so a linked address is followed under both. An address whose script has
// never been spent has no alias yet; the link topic is watched (registered
// before the alias lookup, so no link can fall between the two), and when
// the first spend writes one the alias is followed from then on, including
// the rows that spend wrote. Each key is followed once, however many times
// its link is rewritten or by which encoding it was reached.
func ListenAddrSeenTxs(ctx context.Context, addrs []wallet.LockAddr) (<-chan *addr.SeenTx, error) {
	ctx, cancel := context.WithCancel(ctx)
	watch, err := newLinkWatch(ctx, p2shAddrs(addrs))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error watching p2sh links for address subscription; %w", err)
	}
	aliases, err := attach.GetP2shAliases(ctx, p2shAddrs(addrs))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error getting p2sh aliases for address subscription; %w", err)
	}
	keys := append([]wallet.LockAddr{}, addrs...)
	for _, alias := range aliases {
		keys = append(keys, alias)
	}
	seenTxChan, err := addr.ListenAddrSeenTxs(ctx, watch.followNew(keys))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error getting addr seen txs listener; %w", err)
	}
	var out = make(chan *addr.SeenTx)
	// Rows from aliases that appear after the start are funneled here.
	var lateAliasChan = make(chan *addr.SeenTx)
	go func() {
		defer close(out)
		defer cancel() // stops every listener opened for this subscription
		var delivered = make(map[[32]byte]struct{})
		forward := func(seenTx *addr.SeenTx) bool {
			if _, ok := delivered[seenTx.TxHash]; ok {
				return true
			}
			delivered[seenTx.TxHash] = struct{}{}
			select {
			case <-ctx.Done():
				return false
			case out <- seenTx:
				return true
			}
		}
		for {
			select {
			case <-ctx.Done():
				return
			case seenTx, ok := <-seenTxChan:
				if !ok || !forward(seenTx) {
					return
				}
			case seenTx, ok := <-lateAliasChan:
				if !ok || !forward(seenTx) {
					return
				}
			case link, ok := <-watch.Aliases:
				if !ok {
					return
				}
				alias := link.Other
				if !watch.Follow(alias) {
					continue
				}
				// Listen (registered on return) before reading, so a row
				// written between the two is not missed; one caught by both
				// is deduplicated above. The alias key is new (its link was
				// just written for the first time), so its history is small.
				aliasChan, err := addr.ListenAddrSeenTxs(ctx, []wallet.LockAddr{alias})
				if err != nil {
					log.Printf("error listening to new p2sh alias for address subscription; %v", err)
					return
				}
				go func() {
					// The alias stream ends only with the subscription or
					// with a failed shard stream; either way the subscription
					// is over (cancel ends the main loop, which closes out).
					defer cancel()
					for {
						select {
						case <-ctx.Done():
							return
						case seenTx, ok := <-aliasChan:
							if !ok {
								return
							}
							select {
							case <-ctx.Done():
								return
							case lateAliasChan <- seenTx:
							}
						}
					}
				}()
				seenTxs, err := addr.GetSeenTxs(ctx, alias, nil, 0)
				if err != nil {
					log.Printf("error getting new p2sh alias seen txs for address subscription; %v", err)
					return
				}
				for _, seenTx := range seenTxs {
					if !forward(seenTx) {
						return
					}
				}
			}
		}
	}()
	return out, nil
}
