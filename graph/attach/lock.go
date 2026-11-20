package attach

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Lock struct {
	base
	Locks []*model.Lock
	// Aliases maps a P2SH address (either width) to the other encoding of
	// the same redeem script, for the locks that have one. A lock's history
	// is the union of the rows under both keys: outputs are recorded under
	// the encoding the lock script used, spends always under the P2SH20 one.
	Aliases map[wallet.LockAddr]wallet.LockAddr
}

func ToLocks(ctx context.Context, fields []Field, locks []*model.Lock) error {
	t := Lock{
		base:  base{Ctx: ctx, Fields: fields},
		Locks: locks,
	}
	if t.HasField([]string{"profile", "txs"}) {
		if err := t.attachAliases(); err != nil {
			return fmt.Errorf("error getting p2sh aliases for locks; %w", err)
		}
	}
	t.Wait.Add(2)
	go t.AttachProfiles()
	go t.AttachTxs()
	t.Wait.Wait()
	if len(t.Errors) > 0 {
		return fmt.Errorf("error attaching details to locks; %w", t.Errors[0])
	}
	return nil
}

func (l *Lock) attachAliases() error {
	l.Mutex.Lock()
	var p2shAddrs []wallet.LockAddr
	for _, lock := range l.Locks {
		if lock.Address.LockAddr().IsP2SH() {
			p2shAddrs = append(p2shAddrs, lock.Address.LockAddr())
		}
	}
	l.Mutex.Unlock()
	aliases, err := GetP2shAliases(l.Ctx, p2shAddrs)
	if err != nil {
		return err
	}
	l.Mutex.Lock()
	l.Aliases = aliases
	l.Mutex.Unlock()
	return nil
}

// GetP2shAliases returns the other encoding for each P2SH address that has
// a link row; addresses without one (P2PKH, or a script never yet spent) are
// absent from the map.
func GetP2shAliases(ctx context.Context, addrs []wallet.LockAddr) (map[wallet.LockAddr]wallet.LockAddr, error) {
	var aliases = make(map[wallet.LockAddr]wallet.LockAddr)
	if len(addrs) == 0 {
		return aliases, nil
	}
	links, err := addr.GetP2shLinks(ctx, addrs)
	if err != nil {
		return nil, fmt.Errorf("error getting p2sh links; %w", err)
	}
	for _, link := range links {
		aliases[link.Addr] = link.Other
	}
	return aliases, nil
}

// MemoAddresses resolves requested addresses to the 20-byte keys the memo
// topics use. A P2SH32 address's memo activity is recorded under its P2SH20
// alias (the key) but is reported back under the address the client asked
// for. Each requested address keeps its own entry, and a key requested by
// more than one form (a P2SH32 address and its alias) reports under every
// one of them.
type MemoAddresses struct {
	// Entries hold one item per requested address, in request order.
	Entries []MemoAddress
	// Keys are the distinct memo keys of the entries that have one, in
	// first-request order.
	Keys [][25]byte
	// Unlinked are the distinct P2SH32 addresses with no alias yet. Their
	// script has never been spent, so they have no memo activity; a
	// subscription follows them once their link appears.
	Unlinked []wallet.LockAddr
	mu       sync.Mutex
	forms    map[[25]byte][]model.Address
}

// MemoAddress is one requested address and its memo key.
type MemoAddress struct {
	Address model.Address
	// Key is the memo key; HasKey is false for an unlinked P2SH32 address.
	Key    [25]byte
	HasKey bool
}

// ResolveMemoAddresses looks up the memo key of each address: a 20-byte
// address is its own key, a P2SH32 address resolves to its alias.
func ResolveMemoAddresses(ctx context.Context, addresses []model.Address) (*MemoAddresses, error) {
	var m = &MemoAddresses{forms: make(map[[25]byte][]model.Address)}
	var p2sh32Addrs []wallet.LockAddr
	for _, address := range addresses {
		if address.IsP2SH32() {
			p2sh32Addrs = append(p2sh32Addrs, address.LockAddr())
		}
	}
	aliases, err := GetP2shAliases(ctx, p2sh32Addrs)
	if err != nil {
		return nil, err
	}
	var seenKeys = make(map[[25]byte]struct{})
	var seenUnlinked = make(map[wallet.LockAddr]struct{})
	for _, address := range addresses {
		var entry = MemoAddress{Address: address}
		if !address.IsP2SH32() {
			entry.Key, entry.HasKey = address.Addr(), true
		} else if alias, ok := aliases[address.LockAddr()]; ok {
			entry.Key, entry.HasKey = alias.Addr(), true
		} else if _, dup := seenUnlinked[address.LockAddr()]; !dup {
			seenUnlinked[address.LockAddr()] = struct{}{}
			m.Unlinked = append(m.Unlinked, address.LockAddr())
		}
		if entry.HasKey {
			if _, dup := seenKeys[entry.Key]; !dup {
				seenKeys[entry.Key] = struct{}{}
				m.Keys = append(m.Keys, entry.Key)
			}
			m.AddAlias(entry.Key, address.LockAddr())
		}
		m.Entries = append(m.Entries, entry)
	}
	return m, nil
}

// AddAlias records requested as a form that activity under key is reported
// by, e.g. the P2SH32 address of a link that appeared after the lookup. A
// form already recorded is not repeated.
func (m *MemoAddresses) AddAlias(key [25]byte, requested wallet.LockAddr) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, form := range m.forms[key] {
		if form.LockAddr() == requested {
			return
		}
	}
	m.forms[key] = append(m.forms[key], model.Address(requested))
}

// KeyEntries returns one entry per form key was requested by, for
// attaching a profile under each.
func (m *MemoAddresses) KeyEntries(key [25]byte) []MemoAddress {
	addresses := m.Addresses(key)
	var entries = make([]MemoAddress, len(addresses))
	for i, address := range addresses {
		entries[i] = MemoAddress{Address: address, Key: key, HasKey: true}
	}
	return entries
}

// Addresses returns the forms key was requested by: the key's own 20-byte
// address, the P2SH32 address it is the alias of, or both. A key requested
// by neither (not expected) is reported as itself.
func (m *MemoAddresses) Addresses(key [25]byte) []model.Address {
	m.mu.Lock()
	defer m.mu.Unlock()
	if forms := m.forms[key]; len(forms) > 0 {
		return append([]model.Address{}, forms...)
	}
	return []model.Address{model.NewAddress(key)}
}

func (l *Lock) GetLockAddrs() [][25]byte {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	var lockAddrs [][25]byte
	for _, lock := range l.Locks {
		lockAddrs = append(lockAddrs, lock.Address.LockAddr().Addr())
	}
	return lockAddrs
}

// lockKeys returns the seen-tx keys a lock's history lives under: its own
// address and, when linked, the other encoding of its redeem script.
func (l *Lock) lockKeys(lock *model.Lock) []wallet.LockAddr {
	keys := []wallet.LockAddr{lock.Address.LockAddr()}
	if alias, ok := l.Aliases[lock.Address.LockAddr()]; ok {
		keys = append(keys, alias)
	}
	return keys
}

func (l *Lock) AttachProfiles() {
	defer l.Wait.Done()
	if !l.HasField([]string{"profile"}) {
		return
	}
	// Memo activity is keyed by the 20-byte form, so a P2SH32 lock's
	// profile is that of its alias, reported under the lock's own address;
	// without an alias it is simply empty, with no memo read through a zero
	// key. Locks sharing a key have it attached once.
	l.Mutex.Lock()
	var entries = make([]MemoAddress, len(l.Locks))
	for i, lock := range l.Locks {
		entries[i] = MemoAddress{Address: lock.Address}
		if !lock.Address.IsP2SH32() {
			entries[i].Key, entries[i].HasKey = lock.Address.Addr(), true
		} else if alias, ok := l.Aliases[lock.Address.LockAddr()]; ok {
			entries[i].Key, entries[i].HasKey = alias.Addr(), true
		}
	}
	l.Mutex.Unlock()
	profiles, err := GetMemoProfilesForEntries(l.Ctx, GetPrefixFields(l.Fields, "profile"), entries)
	if err != nil {
		l.AddError(fmt.Errorf("error attaching to lock profiles; %w", err))
		return
	}
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	for i, lock := range l.Locks {
		lock.Profile = profiles[i]
	}
}

func (l *Lock) AttachTxs() {
	defer l.Wait.Done()
	if !l.HasField([]string{"txs"}) {
		return
	}
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	txsField := l.Fields.GetField("txs")
	startDate, _ := model.UnmarshalDate(txsField.Arguments["start"])
	startTx, _ := txsField.Arguments["tx"].(string)
	limit, _ := model.UnmarshalUint32(txsField.Arguments["limit"])
	var startSuffix []byte
	if time.Time(startDate).After(memo.GetGenesisTime()) {
		startSuffix = jutil.GetTimeByteNanoBig(time.Time(startDate))
		if len(startTx) > 0 {
			txHash, err := chainhash.NewHashFromStr(startTx)
			if err != nil {
				l.AddError(fmt.Errorf("error decoding start hash for lock txs resolver; %w", err))
				return
			}
			startSuffix = append(startSuffix, jutil.ByteReverse(txHash[:])...)
		}
	}
	var allTxs []*model.Tx
	for _, lock := range l.Locks {
		var seenTxs []*addr.SeenTx
		for _, key := range l.lockKeys(lock) {
			var startUid []byte
			if len(startSuffix) > 0 {
				startUid = jutil.CombineBytes(key.Bytes(), startSuffix)
			}
			keySeenTxs, err := addr.GetSeenTxs(l.Ctx, key, startUid, uint32(limit))
			if err != nil {
				l.AddError(fmt.Errorf("error getting addr seen txs for lock txs resolver; %w", err))
				return
			}
			seenTxs = append(seenTxs, keySeenTxs...)
		}
		seenTxs = mergeSeenTxs(seenTxs, int(limit))
		lock.Txs = make([]*model.Tx, len(seenTxs))
		for i := range seenTxs {
			lock.Txs[i] = &model.Tx{
				Hash: seenTxs[i].TxHash,
				Seen: model.Date(seenTxs[i].Seen),
			}
		}
		allTxs = append(allTxs, lock.Txs...)
	}
	if err := ToTxs(l.Ctx, txsField.Fields, allTxs); err != nil {
		l.AddError(fmt.Errorf("error attaching to lock transactions; %w", err))
		return
	}
}

// mergeSeenTxs combines rows read under several keys into one history in
// key order (time, then tx hash), dropping a tx seen under more than one key
// (a covenant state transition is an output under one and a spend under the
// other). Each key was read from the same start with the same limit, so the
// first limit rows of the union are exactly the next page.
func mergeSeenTxs(seenTxs []*addr.SeenTx, limit int) []*addr.SeenTx {
	sort.SliceStable(seenTxs, func(i, j int) bool {
		if !seenTxs[i].Seen.Equal(seenTxs[j].Seen) {
			return seenTxs[i].Seen.Before(seenTxs[j].Seen)
		}
		return bytes.Compare(jutil.ByteReverse(seenTxs[i].TxHash[:]), jutil.ByteReverse(seenTxs[j].TxHash[:])) < 0
	})
	var merged = make([]*addr.SeenTx, 0, len(seenTxs))
	var seen = make(map[[32]byte]struct{}, len(seenTxs))
	for _, seenTx := range seenTxs {
		if _, ok := seen[seenTx.TxHash]; ok {
			continue
		}
		seen[seenTx.TxHash] = struct{}{}
		merged = append(merged, seenTx)
		if limit > 0 && len(merged) >= limit {
			break
		}
	}
	return merged
}
