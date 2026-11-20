package tasks

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/db"
	dbMemo "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/db/server"
	"github.com/memocash/index/graph/attach"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/graph/sub"
	act_maint "github.com/memocash/index/node/act/maint"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/dbi"
	"github.com/memocash/index/test/suite"
)

// P2sh32 is an end-to-end test of P2SH32 address support against real queue
// shards. A covenant-style script is locked as P2SH32, spent and re-locked
// (a state transition), then spent again; the saver records outputs under
// the 32-byte key and spends under the P2SH20 alias of the redeem script
// with the link between them, and the read side (lock txs, the Address
// scalar, the address subscription, memo attribution) presents one history
// under either encoding. Finally both backfills are run over txs saved with
// the minimal saver only.
var P2sh32 = suite.Test{
	Name: TestP2sh32,
	Test: func(r *suite.TestRequest) error {
		s := p2sh32State{
			ctx:     context.Background(),
			txSaver: saver.NewCombinedTx(false),
			suite:   r.Suite,
		}
		return s.run()
	},
}

type p2sh32State struct {
	ctx     context.Context
	txSaver *saver.CombinedTx
	suite   *suite.Suite
}

func (s *p2sh32State) save(txs ...*wire.MsgTx) error {
	block := dbi.WireBlockToBlock(memo.GetBlockFromTxs(txs, nil))
	if err := s.txSaver.SaveTxs(s.ctx, block); err != nil {
		return fmt.Errorf("error saving txs; %w", err)
	}
	return nil
}

func lockP2pkh(pkHash []byte) []byte {
	return append(append([]byte{txscript.OP_DUP, txscript.OP_HASH160, txscript.OP_DATA_20}, pkHash...),
		txscript.OP_EQUALVERIFY, txscript.OP_CHECKSIG)
}

func lockP2sh32(redeemScript []byte) []byte {
	_, addr32 := wallet.GetP2shAddrsFromRedeemScript(redeemScript)
	return append(append([]byte{txscript.OP_HASH256, txscript.OP_DATA_32}, addr32[:]...), txscript.OP_EQUAL)
}

func lockP2sh20(redeemScript []byte) []byte {
	p2sh, _ := wallet.GetP2shAddrsFromRedeemScript(redeemScript)
	return append(append([]byte{txscript.OP_HASH160, txscript.OP_DATA_20}, p2sh[1:21]...), txscript.OP_EQUAL)
}

// unlockP2sh builds the unlock script of a P2SH spend: the redeem script as
// the final push (the same shape whether the output was P2SH20 or P2SH32).
func unlockP2sh(redeemScript []byte) []byte {
	unlock, _ := txscript.NewScriptBuilder().AddData(redeemScript).Script()
	return unlock
}

func spendTx(prevHash chainhash.Hash, prevIndex uint32, unlock []byte, outputs ...*wire.TxOut) *wire.MsgTx {
	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(&prevHash, prevIndex), unlock))
	for _, output := range outputs {
		tx.AddTxOut(output)
	}
	return tx
}

func txHashes(txs ...*wire.MsgTx) [][32]byte {
	var hashes = make([][32]byte, len(txs))
	for i, tx := range txs {
		hashes[i] = tx.TxHash()
	}
	return hashes
}

func (s *p2sh32State) lockTxHashes(address model.Address, start *model.Tx, limit uint32) ([][32]byte, error) {
	var txsField = attach.Field{Name: "txs", Arguments: map[string]interface{}{}, Fields: []attach.Field{{Name: "hash"}}}
	if start != nil {
		txsField.Arguments["start"] = time.Time(start.Seen).Format(time.RFC3339Nano)
		txsField.Arguments["tx"] = chainhash.Hash(start.Hash).String()
	}
	if limit > 0 {
		txsField.Arguments["limit"] = int(limit)
	}
	lock := &model.Lock{Address: address}
	if err := attach.ToLocks(s.ctx, []attach.Field{txsField}, []*model.Lock{lock}); err != nil {
		return nil, fmt.Errorf("error attaching lock txs for %s; %w", address, err)
	}
	var hashes = make([][32]byte, len(lock.Txs))
	for i, tx := range lock.Txs {
		hashes[i] = tx.Hash
	}
	return hashes, nil
}

func sameHashes(got, want [][32]byte) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func hashStrings(hashes [][32]byte) []string {
	var strs = make([]string, len(hashes))
	for i, hash := range hashes {
		strs[i] = chainhash.Hash(hash).String()
	}
	return strs
}

func (s *p2sh32State) run() error {
	fundingTx, err := test_tx.GetFundingTx(test_tx.Address1, 1e6)
	if err != nil {
		return fmt.Errorf("error getting funding tx; %w", err)
	}
	if err := s.save(fundingTx.MsgTx); err != nil {
		return err
	}
	// The saver does not verify scripts, so the redeem scripts can be trivial.
	redeem := []byte{txscript.OP_1}
	other := []byte{txscript.OP_2}
	p2shAddr, addr32 := wallet.GetP2shAddrsFromRedeemScript(redeem)
	address32 := model.Address(addr32.LockAddr())
	aliasAddress := model.Address(p2shAddr.LockAddr())

	// A: lock the script as P2SH32, another as P2SH20, plus P2PKH change.
	txA := spendTx(fundingTx.MsgTx.TxHash(), 0, nil,
		wire.NewTxOut(10000, lockP2sh32(redeem)),
		wire.NewTxOut(10000, lockP2sh20(other)),
		wire.NewTxOut(900000, lockP2pkh(test_tx.Address1pkHash)))
	if err := s.save(txA); err != nil {
		return err
	}
	hashA := txA.TxHash()
	seenTxs, err := addr.GetSeenTxs(s.ctx, addr32.LockAddr(), nil, 0)
	if err != nil {
		return fmt.Errorf("error getting p2sh32 seen txs; %w", err)
	}
	if len(seenTxs) != 1 || seenTxs[0].TxHash != hashA || seenTxs[0].Addr != addr32.LockAddr() {
		return fmt.Errorf("error expected one p2sh32 seen tx for A, got %+v", seenTxs)
	}
	links, err := addr.GetP2shLinks(s.ctx, []wallet.LockAddr{addr32.LockAddr()})
	if err != nil {
		return fmt.Errorf("error getting p2sh links; %w", err)
	}
	if len(links) != 0 {
		return fmt.Errorf("error expected no link before first spend, got %+v", links)
	}
	log.Printf("✓ p2sh32 output recorded under its 32-byte key, no link yet: %s\n", addr32)

	// Profiles query: the unlinked P2SH32 address still gets its own (empty)
	// profile, in its requested position and form, with the required lock
	// under that form carrying the outputs it has already received.
	lockField := attach.Field{Name: "lock", Fields: []attach.Field{{Name: "address"}, {Name: "txs", Arguments: map[string]interface{}{}, Fields: []attach.Field{{Name: "hash"}}}}}
	postsFields := attach.Fields{{Name: "posts", Arguments: map[string]interface{}{}, Fields: []attach.Field{{Name: "tx_hash"}}}, lockField}
	pkhAddress := model.NewAddress(*wallet.GetAddrFromPkHash(test_tx.Address1pkHash))
	profiles, err := attach.GetMemoProfiles(s.ctx, postsFields, []model.Address{address32, pkhAddress})
	if err != nil {
		return fmt.Errorf("error getting profiles with unlinked p2sh32 address; %w", err)
	}
	if len(profiles) != 2 || profiles[0].Address != address32 || profiles[1].Address != pkhAddress || len(profiles[0].Posts) != 0 {
		return fmt.Errorf("error profiles with unlinked p2sh32 address: %s", profileStrings(profiles))
	}
	for i, profile := range profiles {
		if profile.Lock == nil || profile.Lock.Address != profile.Address {
			return fmt.Errorf("error profile %d lock missing or under another address: %+v", i, profile.Lock)
		}
	}
	if got := modelTxHashes(profiles[0].Lock.Txs); !sameHashes(got, [][32]byte{hashA}) {
		return fmt.Errorf("error unlinked p2sh32 profile lock txs: %v, want %v", hashStrings(got), hashStrings([][32]byte{hashA}))
	}
	log.Printf("✓ profiles query returns an empty profile with its lock for the unlinked p2sh32 address\n")

	// GraphQL: lock.address is a CashAddr for P2SH32 and base58 for P2SH20,
	// and either string parses back to the same address.
	outputs := []*model.TxOutput{{Hash: model.Hash(hashA), Index: 0}, {Hash: model.Hash(hashA), Index: 1}, {Hash: model.Hash(hashA), Index: 2}}
	fields := []attach.Field{{Name: "script"}, {Name: "lock", Fields: []attach.Field{{Name: "address"}}}}
	if err := attach.ToOutputs(s.ctx, fields, outputs); err != nil {
		return fmt.Errorf("error attaching outputs; %w", err)
	}
	otherAddr, _ := wallet.GetP2shAddrsFromRedeemScript(other)
	for i, want := range []string{addr32.String(), otherAddr.String(), test_tx.Address1String} {
		if outputs[i].Lock == nil || outputs[i].Lock.Address.String() != want {
			return fmt.Errorf("error output %d lock mismatch: %+v, want %s", i, outputs[i].Lock, want)
		}
		parsed, err := model.UnmarshalAddress(want)
		if err != nil || parsed != outputs[i].Lock.Address {
			return fmt.Errorf("error address scalar round trip for %s: %+v %v", want, parsed, err)
		}
	}
	// A 256-bit P2SH CashAddr: "bitcoincash:" + 53 payload chars + 8 checksum chars.
	if str := outputs[0].Lock.Address.String(); len(str) != 73 || str[:13] != "bitcoincash:p" {
		return fmt.Errorf("error p2sh32 address is not a 256-bit p2sh cash address: %s", str)
	}
	log.Printf("✓ graphql lock addresses: %s, %s, %s\n", outputs[0].Lock.Address, outputs[1].Lock.Address, outputs[2].Lock.Address)

	// Subscribe to the P2SH32 address before its script has ever been spent,
	// so no alias exists yet; the subscription must still see the spends.
	subCtx, cancelSub := context.WithCancel(s.ctx)
	defer cancelSub()
	seenChan, err := sub.ListenAddrSeenTxs(subCtx, []wallet.LockAddr{addr32.LockAddr()})
	if err != nil {
		return fmt.Errorf("error listening to p2sh32 address; %w", err)
	}
	// A second subscription names both encodings of the same script: each
	// tx must still arrive once, and the link written by the first spend
	// must not attach another listener for a key it already follows.
	bothChan, err := sub.ListenAddrSeenTxs(subCtx, []wallet.LockAddr{addr32.LockAddr(), p2shAddr.LockAddr()})
	if err != nil {
		return fmt.Errorf("error listening to both p2sh encodings; %w", err)
	}
	expectSeen := func(name string, want [32]byte) error {
		if err := expectSeenTx(seenChan, "p2sh32 subscription", name, want); err != nil {
			return err
		}
		return expectSeenTx(bothChan, "both-encodings subscription", name, want)
	}

	// B: spend the P2SH32 output and re-lock the same script (a covenant
	// state transition): recorded as an output under the 32-byte key and as
	// a spend under the alias, and the link is written.
	txB := spendTx(hashA, 0, unlockP2sh(redeem),
		wire.NewTxOut(9000, lockP2sh32(redeem)),
		wire.NewTxOut(900, lockP2pkh(test_tx.Address2pkHash)))
	if err := s.save(txB); err != nil {
		return err
	}
	hashB := txB.TxHash()
	if err := expectSeen("B", hashB); err != nil {
		return err
	}
	links, err = addr.GetP2shLinks(s.ctx, []wallet.LockAddr{addr32.LockAddr(), p2shAddr.LockAddr()})
	if err != nil {
		return fmt.Errorf("error getting p2sh links after spend; %w", err)
	}
	if len(links) != 2 {
		return fmt.Errorf("error expected links in both directions, got %+v", links)
	}
	for _, link := range links {
		if (link.Addr == addr32.LockAddr()) != (link.Other == p2shAddr.LockAddr()) || link.Addr == link.Other {
			return fmt.Errorf("error unexpected link %+v", *link)
		}
	}
	aliasSeenTxs, err := addr.GetSeenTxs(s.ctx, p2shAddr.LockAddr(), nil, 0)
	if err != nil {
		return fmt.Errorf("error getting alias seen txs; %w", err)
	}
	if len(aliasSeenTxs) != 1 || aliasSeenTxs[0].TxHash != hashB {
		return fmt.Errorf("error expected the spend under the alias only, got %+v", aliasSeenTxs)
	}
	log.Printf("✓ first spend wrote the link and recorded the spend under the alias %s\n", p2shAddr)

	// C: spend again, this time to P2PKH only, so C is only under the alias.
	txC := spendTx(hashB, 0, unlockP2sh(redeem), wire.NewTxOut(8000, lockP2pkh(test_tx.Address3pkHash)))
	if err := s.save(txC); err != nil {
		return err
	}
	hashC := txC.TxHash()
	if err := expectSeen("C", hashC); err != nil {
		return err
	}
	log.Printf("✓ subscriptions delivered the state transition and the final spend once each\n")
	// C arrived under the alias only, so the alias listener is attached.
	// Every spend rewrites the link; later spends must not attach another.
	listenersAfterC := server.ListenCount()

	// Memo: a post signed by the script is attributed to the alias, and the
	// P2SH32 address resolves to it for memo lookups.
	postScript, err := script.Post{Message: "hello from a covenant"}.Get()
	if err != nil {
		return fmt.Errorf("error building post script; %w", err)
	}
	txE := spendTx(hashC, 7, unlockP2sh(redeem), wire.NewTxOut(0, postScript))
	if err := s.save(txE); err != nil {
		return err
	}
	hashE := txE.TxHash()
	if err := expectSeen("E", hashE); err != nil {
		return err
	}
	if listeners := server.ListenCount(); listeners != listenersAfterC {
		return fmt.Errorf("error listeners grew from %d to %d across a repeated spend of the same script",
			listenersAfterC, listeners)
	}
	log.Printf("✓ repeated spends of the same script attach no further listeners (%d open)\n", listenersAfterC)
	post, err := dbMemo.GetPost(s.ctx, hashE)
	if err != nil || post == nil {
		return fmt.Errorf("error getting memo post; %v %+v", err, post)
	}
	if post.Addr != p2shAddr {
		return fmt.Errorf("error memo post attributed to %s, want alias %s", wallet.Addr(post.Addr), p2shAddr)
	}
	address1 := *wallet.GetAddrFromPkHash(test_tx.Address1pkHash)
	address1Model := model.NewAddress(address1)
	memoAddrs, err := attach.ResolveMemoAddresses(s.ctx, []model.Address{address32, address1Model})
	if err != nil {
		return fmt.Errorf("error resolving memo addresses; %w", err)
	}
	if len(memoAddrs.Keys) != 2 || memoAddrs.Keys[0] != p2shAddr || memoAddrs.Keys[1] != address1 || len(memoAddrs.Unlinked) != 0 {
		return fmt.Errorf("error resolved memo addresses mismatch: %v, unlinked %v", memoAddrs.Keys, memoAddrs.Unlinked)
	}
	// Activity under the alias is reported under the address asked for.
	if forms := memoAddrs.Addresses(p2shAddr); len(forms) != 1 || forms[0] != address32 {
		return fmt.Errorf("error alias reported under %v, want %s", forms, address32)
	}
	if forms := memoAddrs.Addresses(address1); len(forms) != 1 || forms[0] != address1Model {
		return fmt.Errorf("error p2pkh reported under %v, want %s", forms, address1Model)
	}
	// Both forms of one script: one key, two entries, reported under both.
	bothAddrs, err := attach.ResolveMemoAddresses(s.ctx, []model.Address{aliasAddress, address32})
	if err != nil {
		return fmt.Errorf("error resolving both forms; %w", err)
	}
	if len(bothAddrs.Keys) != 1 || len(bothAddrs.Entries) != 2 || bothAddrs.Entries[0].Address != aliasAddress ||
		bothAddrs.Entries[1].Address != address32 || len(bothAddrs.Addresses(p2shAddr)) != 2 {
		return fmt.Errorf("error resolving both forms: keys %v, entries %+v", bothAddrs.Keys, bothAddrs.Entries)
	}
	log.Printf("✓ memo post by the script attributed to the alias; p2sh32 resolves to it\n")

	// Profiles query with both forms, in either order: each gets its own
	// profile under the form it was asked for, with the post exactly once,
	// directly and through the nested lock profile (which shares the memo
	// key between the two locks), within a page limit of one.
	nestedPosts := attach.Field{Name: "posts", Arguments: map[string]interface{}{"limit": 1}, Fields: []attach.Field{{Name: "tx_hash"}}}
	nestedLock := attach.Field{Name: "lock", Fields: append(append([]attach.Field{}, lockField.Fields...),
		attach.Field{Name: "profile", Fields: []attach.Field{{Name: "address"}, nestedPosts}})}
	for _, addresses := range [][]model.Address{{address32, aliasAddress}, {aliasAddress, address32}} {
		profiles, err := attach.GetMemoProfiles(s.ctx, attach.Fields{postsFields[0], nestedLock}, addresses)
		if err != nil {
			return fmt.Errorf("error getting profiles for both forms; %w", err)
		}
		if len(profiles) != 2 {
			return fmt.Errorf("error expected 2 profiles for both forms, got %s", profileStrings(profiles))
		}
		for i, profile := range profiles {
			if profile.Address != addresses[i] || len(profile.Posts) != 1 || profile.Posts[0].TxHash != model.Hash(hashE) {
				return fmt.Errorf("error profile %d for both forms: %s", i, profileStrings(profiles))
			}
			if profile.Lock == nil || profile.Lock.Address != addresses[i] || !sameHashes(modelTxHashes(profile.Lock.Txs), [][32]byte{hashA, hashB, hashC, hashE}) {
				return fmt.Errorf("error profile %d lock for both forms: %+v", i, profile.Lock)
			}
			nested := profile.Lock.Profile
			if nested == nil || nested.Address != addresses[i] || len(nested.Posts) != 1 || nested.Posts[0].TxHash != model.Hash(hashE) {
				return fmt.Errorf("error profile %d nested lock profile for both forms: %+v", i, nested)
			}
		}
	}
	log.Printf("✓ profiles query keeps each requested form, with the post once each, directly and through the lock\n")

	// History: one merged, deduplicated history under either encoding.
	want := [][32]byte{hashA, hashB, hashC, hashE}
	for _, address := range []model.Address{address32, aliasAddress} {
		got, err := s.lockTxHashes(address, nil, 0)
		if err != nil {
			return err
		}
		if !sameHashes(got, want) {
			return fmt.Errorf("error history for %s: %v, want %v", address, hashStrings(got), hashStrings(want))
		}
	}
	page, err := s.lockTxHashes(address32, nil, 2)
	if err != nil {
		return err
	}
	if !sameHashes(page, want[:2]) {
		return fmt.Errorf("error first page: %v, want %v", hashStrings(page), hashStrings(want[:2]))
	}
	cSeen, err := addr.GetSeenTxs(s.ctx, p2shAddr.LockAddr(), nil, 0)
	if err != nil {
		return err
	}
	var startC *model.Tx
	for _, seenTx := range cSeen {
		if seenTx.TxHash == hashC {
			startC = &model.Tx{Hash: model.Hash(hashC), Seen: model.Date(seenTx.Seen)}
		}
	}
	page, err = s.lockTxHashes(address32, startC, 0)
	if err != nil {
		return err
	}
	if !sameHashes(page, want[2:]) {
		return fmt.Errorf("error page from C: %v, want %v", hashStrings(page), hashStrings(want[2:]))
	}
	// The P2SH20 script that was never spent has no alias and a plain history.
	got, err := s.lockTxHashes(model.Address(otherAddr.LockAddr()), nil, 0)
	if err != nil {
		return err
	}
	if !sameHashes(got, [][32]byte{hashA}) {
		return fmt.Errorf("error unspent p2sh20 history: %v", hashStrings(got))
	}
	log.Printf("✓ lock txs merge both keys, deduplicate the transition and page correctly\n")
	cancelSub()
	memoAddr32, err := s.runMemoSubscriptions(hashA)
	if err != nil {
		return err
	}
	if err := s.runRegistrationAck(); err != nil {
		return err
	}
	if err := s.runCancellation(addr32, memoAddr32); err != nil {
		return err
	}
	if err := s.runBackfill(hashC); err != nil {
		return err
	}
	// Last, since it stops a shard.
	return s.runShardLoss()
}

// expectSeenTx waits for exactly one tx on a subscription and then a short
// silence.
func expectSeenTx(seenChan <-chan *addr.SeenTx, subName, name string, want [32]byte) error {
	select {
	case seenTx, ok := <-seenChan:
		if !ok {
			return fmt.Errorf("error %s closed waiting for %s", subName, name)
		}
		if seenTx.TxHash != want {
			return fmt.Errorf("error %s got %s waiting for %s", subName, chainhash.Hash(seenTx.TxHash), name)
		}
	case <-time.After(10 * time.Second):
		return fmt.Errorf("error %s timed out waiting for %s", subName, name)
	}
	select {
	case seenTx := <-seenChan:
		return fmt.Errorf("error %s delivered an extra tx after %s: %s", subName, name, chainhash.Hash(seenTx.TxHash))
	case <-time.After(500 * time.Millisecond):
	}
	return nil
}

// runMemoSubscriptions opens profile and room follow subscriptions for a
// P2SH32 address before its script has ever been spent, so no alias exists
// yet, then spends it with memo actions. The first spend's actions and the
// later ones must all be delivered, under the alias the actions are
// attributed to.
func (s *p2sh32State) runMemoSubscriptions(prevHash chainhash.Hash) (wallet.Addr32, error) {
	redeem := []byte{txscript.OP_5}
	p2shAddr, addr32 := wallet.GetP2shAddrsFromRedeemScript(redeem)
	if err := s.memoSubscriptions(prevHash, redeem, p2shAddr, addr32); err != nil {
		return wallet.Addr32{}, err
	}
	return addr32, nil
}

func (s *p2sh32State) memoSubscriptions(prevHash chainhash.Hash, redeem []byte, p2shAddr wallet.Addr, addr32 wallet.Addr32) error {
	address32 := model.Address(addr32.LockAddr())
	// A second script whose very first spend is the room action, so the
	// room follow subscription must reconcile a row saved in the same save
	// as the link it learns the alias from.
	redeemB := []byte{txscript.OP_8}
	_, addr32B := wallet.GetP2shAddrsFromRedeemScript(redeemB)
	address32B := model.Address(addr32B.LockAddr())
	tx0 := spendTx(prevHash, 11, nil, wire.NewTxOut(10000, lockP2sh32(redeem)), wire.NewTxOut(10000, lockP2sh32(redeemB)))
	if err := s.save(tx0); err != nil {
		return err
	}
	subCtx, cancelSub := context.WithCancel(s.ctx)
	defer cancelSub()
	nameFields := attach.Fields{{Name: "name", Fields: []attach.Field{{Name: "name"}}}}
	lockAddressField := attach.Field{Name: "lock", Fields: []attach.Field{{Name: "address"}}}
	profileChan, err := new(sub.Profile).Listen(subCtx, []model.Address{address32}, nameFields)
	if err != nil {
		return fmt.Errorf("error listening to profile of unspent p2sh32 address; %w", err)
	}
	// A subscription naming both forms of the (still unlinked) script: the
	// alias is followed directly, and once the first spend links them each
	// name event must arrive under both forms.
	aliasAddress := model.NewAddress(p2shAddr)
	profileChanBoth, err := new(sub.Profile).Listen(subCtx, []model.Address{address32, aliasAddress}, append(nameFields, lockAddressField))
	if err != nil {
		return fmt.Errorf("error listening to profile of both forms; %w", err)
	}
	roomChan, err := new(sub.RoomFollow).Listen(subCtx, []model.Address{address32, address32B}, attach.Fields{lockAddressField})
	if err != nil {
		return fmt.Errorf("error listening to room follows of unspent p2sh32 address; %w", err)
	}
	// A profile subscription whose only address is unlinked, asking for the
	// follow fields too: it must stay open with nothing to listen to yet,
	// then follow the alias once the first spend links it.
	followFields := attach.Fields{{Name: "following"}, {Name: "followers"}, {Name: "name", Fields: []attach.Field{{Name: "name"}}}}
	profileChanB, err := new(sub.Profile).Listen(subCtx, []model.Address{address32B}, followFields)
	if err != nil {
		return fmt.Errorf("error listening to profile of sole unspent p2sh32 address; %w", err)
	}
	select {
	case profile, ok := <-profileChanB:
		return fmt.Errorf("error sole unlinked profile subscription ended early: %v %+v", ok, profile)
	case <-time.After(300 * time.Millisecond):
	}
	// Both subscriptions resolved nothing: the script has no alias yet.
	if memoAddrs, err := attach.ResolveMemoAddresses(s.ctx, []model.Address{address32}); err != nil {
		return err
	} else if len(memoAddrs.Unlinked) != 1 || memoAddrs.Unlinked[0] != addr32.LockAddr() {
		return fmt.Errorf("error expected the unspent p2sh32 address to be unlinked, got %v", memoAddrs.Unlinked)
	}
	// Each profile event is a full snapshot. A new alias is emitted once on
	// attach and again for each memo row its listeners see, so the same
	// snapshot can repeat (or, if the attach read ran before the row was
	// saved, precede the row); a snapshot with any other name is an error.
	var lastName string
	expectName := func(step, want string) error {
		for deadline := time.After(10 * time.Second); ; {
			select {
			case profile, ok := <-profileChan:
				if !ok {
					return fmt.Errorf("error profile subscription closed waiting for %s", step)
				}
				if profile.Address != address32 {
					return fmt.Errorf("error profile for %s reported under %s, want %s", step, profile.Address, address32)
				}
				var name string
				if profile.Name != nil {
					name = profile.Name.Name
				}
				if name == want {
					lastName = want
					return nil
				} else if name != lastName {
					return fmt.Errorf("error profile for %s has name %q, want %q", step, name, want)
				}
			case <-deadline:
				return fmt.Errorf("error profile subscription timed out waiting for %s", step)
			}
		}
	}
	expectRoom := func(step, room string, unfollow bool, txHash [32]byte, wantAddr model.Address) error {
		select {
		case roomFollow, ok := <-roomChan:
			if !ok {
				return fmt.Errorf("error room follow subscription closed waiting for %s", step)
			}
			if roomFollow.Name != room || roomFollow.Unfollow != unfollow || roomFollow.TxHash != model.Hash(txHash) ||
				roomFollow.Address != wantAddr || roomFollow.Lock == nil || roomFollow.Lock.Address != wantAddr {
				return fmt.Errorf("error room follow for %s mismatch: %+v", step, *roomFollow)
			}
		case <-time.After(10 * time.Second):
			return fmt.Errorf("error room follow subscription timed out waiting for %s", step)
		}
		select {
		case roomFollow := <-roomChan:
			return fmt.Errorf("error room follow subscription delivered an extra event after %s: %+v", step, *roomFollow)
		case <-time.After(500 * time.Millisecond):
		}
		return nil
	}

	// First spend: sets a name. Its link and name row are written by the
	// same save, so the subscription has to catch an alias it learns of
	// only now.
	nameScript, err := script.SetName{Name: "covenant one"}.Get()
	if err != nil {
		return err
	}
	tx1 := spendTx(tx0.TxHash(), 0, unlockP2sh(redeem), wire.NewTxOut(9000, lockP2sh32(redeem)), wire.NewTxOut(0, nameScript))
	if err := s.save(tx1); err != nil {
		return err
	}
	if err := expectName("first name", "covenant one"); err != nil {
		return err
	}
	// The both-forms subscription reports the name under each form (a
	// snapshot may repeat; any other name or address is an error).
	seenForms := map[model.Address]bool{}
	for deadline := time.After(10 * time.Second); !seenForms[address32] || !seenForms[aliasAddress]; {
		select {
		case profile, ok := <-profileChanBoth:
			if !ok {
				return fmt.Errorf("error both-forms profile subscription closed")
			}
			if profile.Address != address32 && profile.Address != aliasAddress {
				return fmt.Errorf("error both-forms profile reported under %s", profile.Address)
			}
			if profile.Lock == nil || profile.Lock.Address != profile.Address {
				return fmt.Errorf("error both-forms profile lock missing or under another address: %+v", profile.Lock)
			}
			if profile.Name == nil {
				continue
			} else if profile.Name.Name != "covenant one" {
				return fmt.Errorf("error both-forms profile has name %q", profile.Name.Name)
			}
			seenForms[profile.Address] = true
		case <-deadline:
			return fmt.Errorf("error both-forms profile subscription timed out, seen %v", seenForms)
		}
	}
	log.Printf("✓ profile subscription on both forms reported the name under each\n")
	nameScript2, err := script.SetName{Name: "covenant two"}.Get()
	if err != nil {
		return err
	}
	tx2 := spendTx(tx1.TxHash(), 0, unlockP2sh(redeem), wire.NewTxOut(8000, lockP2sh32(redeem)), wire.NewTxOut(0, nameScript2))
	if err := s.save(tx2); err != nil {
		return err
	}
	if err := expectName("second name", "covenant two"); err != nil {
		return err
	}
	log.Printf("✓ profile subscription opened before the first spend delivered the first and later names\n")

	followScript, err := script.TopicFollow{TopicName: "covenants"}.Get()
	if err != nil {
		return err
	}
	tx3 := spendTx(tx2.TxHash(), 0, unlockP2sh(redeem), wire.NewTxOut(7000, lockP2sh32(redeem)), wire.NewTxOut(0, followScript))
	if err := s.save(tx3); err != nil {
		return err
	}
	if err := expectRoom("room follow", "covenants", false, tx3.TxHash(), address32); err != nil {
		return err
	}
	unfollowScript, err := script.TopicFollow{TopicName: "covenants", Unfollow: true}.Get()
	if err != nil {
		return err
	}
	tx4 := spendTx(tx3.TxHash(), 0, unlockP2sh(redeem), wire.NewTxOut(6000, lockP2sh32(redeem)), wire.NewTxOut(0, unfollowScript))
	if err := s.save(tx4); err != nil {
		return err
	}
	if err := expectRoom("room unfollow", "covenants", true, tx4.TxHash(), address32); err != nil {
		return err
	}
	log.Printf("✓ room follow subscription opened before the first spend delivered the follow and unfollow once each\n")

	// The second script's first spend is itself a room follow: its link and
	// its room follow row land in the same save.
	firstFollowScript, err := script.TopicFollow{TopicName: "first spend"}.Get()
	if err != nil {
		return err
	}
	tx5 := spendTx(tx0.TxHash(), 1, unlockP2sh(redeemB), wire.NewTxOut(9000, lockP2sh32(redeemB)), wire.NewTxOut(0, firstFollowScript))
	if err := s.save(tx5); err != nil {
		return err
	}
	if err := expectRoom("first-spend room follow", "first spend", false, tx5.TxHash(), address32B); err != nil {
		return err
	}
	log.Printf("✓ a room follow in a script's first spend is delivered once alongside its link\n")

	// The sole-unlinked profile subscription: the first spend attaches the
	// alias (one snapshot), and a later set name arrives through the name
	// listener attached then.
	expectProfileB := func(step, wantName string) error {
		for deadline := time.After(10 * time.Second); ; {
			select {
			case profile, ok := <-profileChanB:
				if !ok {
					return fmt.Errorf("error sole unlinked profile subscription closed waiting for %s", step)
				}
				if profile.Address != address32B {
					return fmt.Errorf("error profile for %s reported under %s, want %s", step, profile.Address, address32B)
				}
				var name string
				if profile.Name != nil {
					name = profile.Name.Name
				}
				if name == wantName {
					return nil
				} else if name != "" {
					return fmt.Errorf("error profile for %s has name %q, want %q", step, name, wantName)
				}
			case <-deadline:
				return fmt.Errorf("error sole unlinked profile subscription timed out waiting for %s", step)
			}
		}
	}
	if err := expectProfileB("alias snapshot", ""); err != nil {
		return err
	}
	nameScriptB, err := script.SetName{Name: "covenant b"}.Get()
	if err != nil {
		return err
	}
	tx6 := spendTx(tx5.TxHash(), 0, unlockP2sh(redeemB), wire.NewTxOut(8000, lockP2sh32(redeemB)), wire.NewTxOut(0, nameScriptB))
	if err := s.save(tx6); err != nil {
		return err
	}
	if err := expectProfileB("name after link", "covenant b"); err != nil {
		return err
	}
	log.Printf("✓ profile subscription with follow fields on a sole unlinked address stayed open and followed the alias\n")
	return nil
}

// runShardLoss stops one queue shard while the other stays up: a merged
// listener and a subscription on a key held by the stopped shard must end,
// so the client learns it no longer sees every shard and can resubscribe,
// instead of silently missing that shard's rows. Runs last.
func (s *p2sh32State) runShardLoss() error {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()
	merged, err := db.ListenPrefixes(ctx, db.TopicAddrSeenTx, map[uint32][][]byte{0: {[]byte{1}}, 1: {[]byte{2}}})
	if err != nil {
		return fmt.Errorf("error listening on both shards; %w", err)
	}
	// A P2SH32 address whose seen-tx key is held by shard 0.
	var onShard0 wallet.Addr32
	for i := 0; ; i++ {
		_, candidate := wallet.GetP2shAddrsFromRedeemScript([]byte{txscript.OP_9, byte(i)})
		if db.GetShardIdFromByte32(candidate.LockAddr().Bytes()) == 0 {
			onShard0 = candidate
			break
		}
	}
	seenChan, err := sub.ListenAddrSeenTxs(ctx, []wallet.LockAddr{onShard0.LockAddr()})
	if err != nil {
		return fmt.Errorf("error listening to address on shard 0; %w", err)
	}
	// A P2PKH address whose memo keys are held by shard 0, and the unlinked
	// P2SH32 address above, whose only stream is its link watch on shard 0.
	var pkhOnShard0 wallet.Addr
	for i := 0; ; i++ {
		pkHash := append([]byte{byte(i)}, test_tx.Address1pkHash[1:]...)
		candidate := wallet.GetAddrFromPkHash(pkHash)
		if db.GetShardIdFromByte32(candidate[:]) == 0 {
			pkhOnShard0 = *candidate
			break
		}
	}
	pkhAddress, unlinkedAddress := model.NewAddress(pkhOnShard0), model.Address(onShard0.LockAddr())
	nameField := attach.Fields{{Name: "name", Fields: []attach.Field{{Name: "name"}}}}
	roomChan, err := new(sub.RoomFollow).Listen(ctx, []model.Address{pkhAddress}, nil)
	if err != nil {
		return fmt.Errorf("error listening to room follows on shard 0; %w", err)
	}
	profileChan, err := new(sub.Profile).Listen(ctx, []model.Address{pkhAddress}, nameField)
	if err != nil {
		return fmt.Errorf("error listening to profile on shard 0; %w", err)
	}
	watchRoomChan, err := new(sub.RoomFollow).Listen(ctx, []model.Address{unlinkedAddress}, nil)
	if err != nil {
		return fmt.Errorf("error listening to room follows of unlinked address on shard 0; %w", err)
	}
	watchProfileChan, err := new(sub.Profile).Listen(ctx, []model.Address{unlinkedAddress}, nameField)
	if err != nil {
		return fmt.Errorf("error listening to profile of unlinked address on shard 0; %w", err)
	}
	s.suite.Queue0.End()
	for name, out := range map[string]func() (bool, bool){
		"merged listener":                      func() (bool, bool) { _, ok := <-merged; return ok, true },
		"address subscription":                 func() (bool, bool) { _, ok := <-seenChan; return ok, true },
		"room follow subscription":             func() (bool, bool) { _, ok := <-roomChan; return ok, true },
		"profile subscription":                 func() (bool, bool) { _, ok := <-profileChan; return ok, true },
		"link-watch-only room subscription":    func() (bool, bool) { _, ok := <-watchRoomChan; return ok, true },
		"link-watch-only profile subscription": func() (bool, bool) { _, ok := <-watchProfileChan; return ok, true },
	} {
		var done = make(chan bool, 1)
		go func() { ok, _ := out(); done <- ok }()
		select {
		case ok := <-done:
			if ok {
				return fmt.Errorf("error %s delivered an event after a shard stopped", name)
			}
		case <-time.After(5 * time.Second):
			return fmt.Errorf("error %s stayed open after shard 0 stopped while shard 1 is up", name)
		}
	}
	log.Printf("✓ losing one shard ends the merged listener and every subscription kind on it\n")
	return nil
}

// listenerGoroutines counts the goroutines of the subscription chain: the
// graph subscriptions, the item listener wrappers, the shard merger and the
// client streams. All of them must end when a subscription is cancelled.
func listenerGoroutines() int {
	var buf = make([]byte, 1<<20)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	var count int
	for _, stack := range strings.Split(string(buf), "\n\n") {
		for _, frame := range []string{
			"index/graph/sub.",
			"index/db/item/addr.Listen",
			"index/db/item/memo.Listen",
			"index/db/item/db.ListenPrefixes",
			"index/db/client.(*Client).Listen",
		} {
			if strings.Contains(stack, frame) {
				count++
				break
			}
		}
	}
	return count
}

// settledListenerGoroutines waits for the count to stop changing.
func settledListenerGoroutines() int {
	count := listenerGoroutines()
	for stable := 0; stable < 10; {
		time.Sleep(50 * time.Millisecond)
		if next := listenerGoroutines(); next == count {
			stable++
		} else {
			count, stable = next, 0
		}
	}
	return count
}

// runCancellation opens every kind of subscription on linked and unlinked
// P2SH32 addresses, cancels them, and requires every goroutine of the
// listener chain to end: a cancelled subscription must leave nothing behind
// even though the server-side listener count alone cannot show a stuck
// forwarder.
func (s *p2sh32State) runCancellation(linked, memoLinked wallet.Addr32) error {
	before := settledListenerGoroutines()
	allFields := attach.Fields{{Name: "name", Fields: []attach.Field{{Name: "name"}}}, {Name: "profile"}, {Name: "pic"},
		{Name: "following"}, {Name: "followers"}}
	for i := 0; i < 5; i++ {
		_, fresh := wallet.GetP2shAddrsFromRedeemScript([]byte{txscript.OP_7, byte(i)})
		linkedAddress, memoAddress, freshAddress :=
			model.Address(linked.LockAddr()), model.Address(memoLinked.LockAddr()), model.Address(fresh.LockAddr())
		ctx, cancel := context.WithCancel(s.ctx)
		seenChan, err := sub.ListenAddrSeenTxs(ctx, []wallet.LockAddr{linked.LockAddr(), fresh.LockAddr()})
		if err != nil {
			cancel()
			return fmt.Errorf("error listening to addresses for cancellation check; %w", err)
		}
		profileChan, err := new(sub.Profile).Listen(ctx, []model.Address{memoAddress, freshAddress}, allFields)
		if err != nil {
			cancel()
			return fmt.Errorf("error listening to profiles for cancellation check; %w", err)
		}
		roomChan, err := new(sub.RoomFollow).Listen(ctx, []model.Address{memoAddress, linkedAddress, freshAddress}, nil)
		if err != nil {
			cancel()
			return fmt.Errorf("error listening to room follows for cancellation check; %w", err)
		}
		if listenerGoroutines() <= before {
			cancel()
			return fmt.Errorf("error expected open subscriptions to add listener goroutines (baseline %d)", before)
		}
		cancel()
		for name, closed := range map[string]func() bool{
			"address": func() bool { _, ok := <-seenChan; return !ok },
			"profile": func() bool { _, ok := <-profileChan; return !ok },
			"room":    func() bool { _, ok := <-roomChan; return !ok },
		} {
			var done = make(chan bool, 1)
			go func() { done <- closed() }()
			select {
			case ok := <-done:
				if !ok {
					return fmt.Errorf("error %s subscription delivered an event after cancellation (cycle %d)", name, i)
				}
			case <-time.After(5 * time.Second):
				return fmt.Errorf("error %s subscription did not close after cancellation (cycle %d)", name, i)
			}
		}
	}
	for deadline := time.After(5 * time.Second); ; {
		if after := listenerGoroutines(); after <= before {
			log.Printf("✓ cancelled subscriptions leave no listener goroutines behind (%d before, %d after)\n", before, after)
			return nil
		}
		select {
		case <-deadline:
			return fmt.Errorf("error listener goroutines leaked after cancellation: %d before, %d after",
				before, listenerGoroutines())
		case <-time.After(20 * time.Millisecond):
		}
	}
}

// runRegistrationAck checks that a listener is registered on the server by
// the time Listen returns: a row saved immediately afterwards is delivered.
// Repeated so a registration that merely tends to win the race would fail.
func (s *p2sh32State) runRegistrationAck() error {
	for i := 0; i < 25; i++ {
		redeem := []byte{txscript.OP_6, byte(i)}
		p2shAddr, addr32 := wallet.GetP2shAddrsFromRedeemScript(redeem)
		listenCtx, cancel := context.WithCancel(s.ctx)
		linkChan, err := addr.ListenP2shLinks(listenCtx, []wallet.LockAddr{addr32.LockAddr()})
		if err != nil {
			cancel()
			return fmt.Errorf("error listening to p2sh links; %w", err)
		}
		if err := db.Save(addr.NewP2shLinks(p2shAddr, addr32)); err != nil {
			cancel()
			return fmt.Errorf("error saving p2sh links; %w", err)
		}
		select {
		case link, ok := <-linkChan:
			if !ok || link.Other != p2shAddr.LockAddr() {
				cancel()
				return fmt.Errorf("error unexpected link on iteration %d: %v %+v", i, ok, link)
			}
		case <-time.After(5 * time.Second):
			cancel()
			return fmt.Errorf("error link saved right after Listen returned was not delivered (iteration %d)", i)
		}
		cancel()
	}
	log.Printf("✓ a row saved immediately after Listen returns is delivered (25/25)\n")
	return nil
}

// runBackfill saves txs through the minimal saver only (chain rows, no
// address rows), runs both backfills and checks the rows they wrote, then
// checks a cancelled context makes Populate fail promptly.
func (s *p2sh32State) runBackfill(prevHash chainhash.Hash) error {
	redeem2 := []byte{txscript.OP_3}
	redeem3 := []byte{txscript.OP_4}
	p2sh3, addr32of3 := wallet.GetP2shAddrsFromRedeemScript(redeem3)
	_, addr32of2 := wallet.GetP2shAddrsFromRedeemScript(redeem2)
	txD := spendTx(prevHash, 9, unlockP2sh(redeem3), wire.NewTxOut(7000, lockP2sh32(redeem2)))
	block := dbi.WireBlockToBlock(memo.GetBlockFromTxs([]*wire.MsgTx{txD}, nil))
	if err := saver.NewTxMinimal(false).SaveTxs(s.ctx, block); err != nil {
		return fmt.Errorf("error saving backfill tx with minimal saver; %w", err)
	}
	hashD := txD.TxHash()
	seenTxs, err := addr.GetSeenTxs(s.ctx, addr32of2.LockAddr(), nil, 0)
	if err != nil {
		return err
	}
	if len(seenTxs) != 0 {
		return fmt.Errorf("error expected no seen txs before backfill, got %+v", seenTxs)
	}

	outputsPopulate := act_maint.NewPopulateP2sh32(s.ctx, false)
	outputsPopulate.LogInterval = time.Millisecond
	if err := outputsPopulate.Populate(true); err != nil {
		return fmt.Errorf("error running populate p2sh32 outputs; %w", err)
	}
	// A:0, B:0, D:0 and the eight P2SH32 outputs of the memo subscription
	// section are the P2SH32 outputs saved so far.
	if outputsPopulate.Found != 11 {
		return fmt.Errorf("error expected backfill to find 11 p2sh32 outputs, got %d", outputsPopulate.Found)
	}
	seenTxs, err = addr.GetSeenTxs(s.ctx, addr32of2.LockAddr(), nil, 0)
	if err != nil {
		return err
	}
	if len(seenTxs) != 1 || seenTxs[0].TxHash != [32]byte(hashD) {
		return fmt.Errorf("error backfill did not write the p2sh32 seen tx: %+v", seenTxs)
	}
	log.Printf("✓ populate-p2sh32-outputs wrote the missing row (checked %d, found %d, saved %d)\n",
		outputsPopulate.Checked, outputsPopulate.Found, outputsPopulate.Saved)

	linksPopulate := act_maint.NewPopulateP2sh32(s.ctx, true)
	linksPopulate.LogInterval = time.Millisecond
	if err := linksPopulate.Populate(true); err != nil {
		return fmt.Errorf("error running populate p2sh links; %w", err)
	}
	links, err := addr.GetP2shLinks(s.ctx, []wallet.LockAddr{addr32of3.LockAddr(), p2sh3.LockAddr()})
	if err != nil {
		return err
	}
	if len(links) != 2 {
		return fmt.Errorf("error backfill did not write both links for the unsaved spend: %+v", links)
	}
	log.Printf("✓ populate-p2sh-links wrote the missing links (checked %d, p2sh inputs %d, saved %d)\n",
		linksPopulate.Checked, linksPopulate.Found, linksPopulate.Saved)

	cancelled, cancel := context.WithCancel(s.ctx)
	cancel()
	failing := act_maint.NewPopulateP2sh32(cancelled, false)
	failing.LogInterval = time.Millisecond
	var result = make(chan error, 1)
	go func() { result <- failing.Populate(true) }()
	select {
	case err := <-result:
		if err == nil {
			return fmt.Errorf("error expected populate with cancelled context to fail")
		}
		if failing.Saved != 0 {
			return fmt.Errorf("error populate with cancelled context saved %d rows", failing.Saved)
		}
	case <-time.After(10 * time.Second):
		return fmt.Errorf("error populate with cancelled context did not return")
	}
	log.Printf("✓ populate returns promptly on failure\n")
	return nil
}

func profileStrings(profiles []*model.Profile) string {
	var parts []string
	for _, profile := range profiles {
		parts = append(parts, fmt.Sprintf("%s(%d posts)", profile.Address, len(profile.Posts)))
	}
	return strings.Join(parts, ", ")
}

func modelTxHashes(txs []*model.Tx) [][32]byte {
	var hashes = make([][32]byte, len(txs))
	for i, tx := range txs {
		hashes[i] = tx.Hash
	}
	return hashes
}
