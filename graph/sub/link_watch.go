package sub

import (
	"context"
	"fmt"
	"sync"

	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// linkWatch follows the P2SH20/P2SH32 link topic for a set of P2SH addresses
// and reports each distinct link once, the first time it appears (every
// spend rewrites the same link). A subscription registers the watch before
// it looks up existing links, so a link written at any moment is seen
// exactly one way: by the lookup if it was already there, by the watch if it
// arrived later (the watch's listener is registered on the server before
// newLinkWatch returns). The consumer marks the keys it follows with Follow,
// so a link whose other side it already follows (because the client asked
// for that form too) adds no listener, while still telling the consumer the
// two forms belong together.
type linkWatch struct {
	// Aliases delivers each distinct link once; nil when nothing is
	// watched, which blocks forever in a select.
	Aliases  <-chan *addr.P2shLink
	mu       sync.Mutex
	followed map[wallet.LockAddr]struct{}
}

// newLinkWatch registers a link listener for the given P2SH addresses (of
// either width) and returns once it is registered; with no addresses it
// returns a watch that never delivers.
func newLinkWatch(ctx context.Context, p2shAddrs []wallet.LockAddr) (*linkWatch, error) {
	var w = &linkWatch{followed: make(map[wallet.LockAddr]struct{})}
	if len(p2shAddrs) == 0 {
		return w, nil
	}
	linkChan, err := addr.ListenP2shLinks(ctx, p2shAddrs)
	if err != nil {
		return nil, fmt.Errorf("error getting p2sh link listener; %w", err)
	}
	var aliases = make(chan *addr.P2shLink)
	w.Aliases = aliases
	go func() {
		defer close(aliases)
		var seen = make(map[[2]wallet.LockAddr]struct{})
		for {
			select {
			case <-ctx.Done():
				return
			case link, ok := <-linkChan:
				if !ok {
					return
				}
				pair := [2]wallet.LockAddr{link.Addr, link.Other}
				if _, ok := seen[pair]; ok {
					continue
				}
				seen[pair] = struct{}{}
				select {
				case <-ctx.Done():
					return
				case aliases <- link:
				}
			}
		}
	}()
	return w, nil
}

// Follow marks a key as followed and reports whether it was new. The caller
// attaches listeners only for keys that were new, so a key reached by
// request, by lookup or by a link is attached once, whichever came first.
func (w *linkWatch) Follow(key wallet.LockAddr) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.followed[key]; ok {
		return false
	}
	w.followed[key] = struct{}{}
	return true
}

// followNew returns the keys that were not yet followed, marking them.
func (w *linkWatch) followNew(keys []wallet.LockAddr) []wallet.LockAddr {
	var newKeys []wallet.LockAddr
	for _, key := range keys {
		if w.Follow(key) {
			newKeys = append(newKeys, key)
		}
	}
	return newKeys
}

// p2shAddrs returns the P2SH addresses, of either width, among addrs.
func p2shAddrs(addrs []wallet.LockAddr) []wallet.LockAddr {
	var p2sh []wallet.LockAddr
	for _, address := range addrs {
		if address.IsP2SH() {
			p2sh = append(p2sh, address)
		}
	}
	return p2sh
}

// p2sh32Addrs returns the P2SH32 addresses among addrs.
func p2sh32Addrs(addrs []wallet.LockAddr) []wallet.LockAddr {
	var p2sh32 []wallet.LockAddr
	for _, address := range addrs {
		if address.IsP2SH32() {
			p2sh32 = append(p2sh32, address)
		}
	}
	return p2sh32
}
