package tasks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	item_slp "github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/graph/attach"
	"github.com/memocash/index/graph/model"
	act_maint "github.com/memocash/index/node/act/maint"
	"github.com/memocash/index/node/act/slp_validate"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	slp_tx "github.com/memocash/index/ref/bitcoin/tx/slp"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/dbi"
	"github.com/memocash/index/test/suite"
	"log"
	"time"
)

// SlpValidity is an end-to-end test of SLP validity verdicts against real
// queue shards: a token lifecycle (genesis, send, mint) is fed through the
// full saver pipeline, alongside a fake send and fake mint with no valid
// token inputs, out-of-order (child before parent) arrival, the
// slp-validity-sweep index sweep, the block-ordered slp-validity-backfill,
// and NFT1 group/child geneses.
var SlpValidity = suite.Test{
	Name: TestSlpValidity,
	Test: func(r *suite.TestRequest) error {
		s := slpValidityState{
			ctx:     context.Background(),
			txSaver: saver.NewCombinedTx(false),
		}
		return s.run()
	},
}

type slpValidityState struct {
	ctx     context.Context
	txSaver *saver.CombinedTx
}

func (s *slpValidityState) save(txs ...*memo.Tx) error {
	var msgTxs = make([]*wire.MsgTx, len(txs))
	for i := range txs {
		msgTxs[i] = txs[i].MsgTx
	}
	block := dbi.WireBlockToBlock(memo.GetBlockFromTxs(msgTxs, nil))
	if err := s.txSaver.SaveTxs(s.ctx, block); err != nil {
		return fmt.Errorf("error saving txs; %w", err)
	}
	return nil
}

func txHash32(tx *memo.Tx) [32]byte {
	var txHash [32]byte
	copy(txHash[:], tx.GetHash())
	return txHash
}

// rawScript emits fixed script bytes, for fixture txs whose vout 0 must be
// byte-exact (e.g. a malformed SLP message)
type rawScript struct{ pk []byte }

func (r rawScript) Get() ([]byte, error)  { return r.pk, nil }
func (r rawScript) Type() memo.OutputType { return memo.OutputTypeUnknown }

// mine writes block associations for txs at the given height (nonce makes the
// block hash unique) via the minimal saver, which skips the live SLP handler,
// so the block-ordered backfill can visit them without re-validating here.
func (s *slpValidityState) mine(height int64, nonce uint32, txs ...*memo.Tx) error {
	var msgTxs = make([]*wire.MsgTx, len(txs))
	for i := range txs {
		msgTxs[i] = txs[i].MsgTx
	}
	header := wire.BlockHeader{Timestamp: time.Unix(1700000000, 0), Nonce: nonce}
	if err := saver.NewTxMinimal(false).SaveTxs(s.ctx, dbi.WireBlockToBlock(
		memo.GetBlockFromTxs(msgTxs, &header))); err != nil {
		return fmt.Errorf("error saving minimal block at height %d; %w", height, err)
	}
	if err := db.Save([]db.Object{
		&chain.BlockHeight{BlockHash: [32]byte(header.BlockHash()), Height: height},
	}); err != nil {
		return fmt.Errorf("error saving block height %d; %w", height, err)
	}
	return nil
}

// checkStatus asserts a tx's stored verdict; status pending = no row.
func (s *slpValidityState) checkStatus(name string, tx *memo.Tx, status slp_tx.Status, reason slp_tx.Reason) error {
	validities, err := item_slp.GetValidities(s.ctx, [][32]byte{txHash32(tx)})
	if err != nil {
		return fmt.Errorf("error getting validities for %s; %w", name, err)
	}
	if status == slp_tx.StatusPending {
		if len(validities) != 0 {
			return fmt.Errorf("error %s: expected pending (no row), got status %d reason %d",
				name, validities[0].Status, validities[0].Reason)
		}
		log.Printf("✓ %s: PENDING (no verdict row)\n", name)
		return nil
	}
	if len(validities) != 1 {
		return fmt.Errorf("error %s: expected status %d, got no verdict row (pending)", name, status)
	}
	if validities[0].Status != uint8(status) || validities[0].Reason != uint8(reason) {
		return fmt.Errorf("error %s: expected status %d reason %d, got status %d reason %d",
			name, status, reason, validities[0].Status, validities[0].Reason)
	}
	log.Printf("✓ %s: status %d reason %d\n", name, status, reason)
	return nil
}

func (s *slpValidityState) wallet(utxos ...memo.UTXO) build.Wallet {
	return build.Wallet{
		Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: utxos}, test_tx.Address1pkHash),
		Address: test_tx.Address1,
		KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
	}
}

func utxoAt(tx *memo.Tx, index uint32) memo.UTXO {
	return memo.UTXO{Input: memo.TxInput{
		PkScript:     tx.MsgTx.TxOut[index].PkScript,
		PkHash:       test_tx.Address1pkHash,
		PrevOutHash:  tx.GetHash(),
		PrevOutIndex: index,
		Value:        tx.MsgTx.TxOut[index].Value,
	}}
}

func tokenUtxoAt(tx *memo.Tx, index uint32, tokenHash []byte, quantity uint64) memo.UTXO {
	utxo := utxoAt(tx, index)
	utxo.SlpToken = tokenHash
	utxo.SlpQuantity = quantity
	return utxo
}

// bchChange returns the tx's last output, the BCH change in every built tx here.
func bchChange(tx *memo.Tx) memo.UTXO {
	return utxoAt(tx, uint32(len(tx.MsgTx.TxOut)-1))
}

// tokenChange finds the vout carrying the given declared quantity by strict
// parsing the tx's own vout-0 message.
func tokenChange(tx *memo.Tx, tokenHash []byte, quantity uint64) (memo.UTXO, error) {
	msg, err := slp_tx.Parse(tx.MsgTx.TxOut[0].PkScript)
	if err != nil {
		return memo.UTXO{}, fmt.Errorf("error parsing token tx for change; %w", err)
	}
	if msg.Send == nil {
		return memo.UTXO{}, fmt.Errorf("error token tx is not a send")
	}
	for i, q := range msg.Send.Quantities {
		if q == quantity {
			return tokenUtxoAt(tx, uint32(i+1), tokenHash, quantity), nil
		}
	}
	return memo.UTXO{}, fmt.Errorf("error token change quantity %d not found", quantity)
}

func (s *slpValidityState) fund(amount int64) (*memo.Tx, memo.UTXO, error) {
	fundingTx, err := test_tx.GetFundingTx(test_tx.Address1, amount)
	if err != nil {
		return nil, memo.UTXO{}, fmt.Errorf("error getting funding tx; %w", err)
	}
	if err := s.save(fundingTx); err != nil {
		return nil, memo.UTXO{}, fmt.Errorf("error saving funding tx; %w", err)
	}
	return fundingTx, utxoAt(fundingTx, 0), nil
}

func (s *slpValidityState) run() error {
	// Plain BCH funding, ingested first so SLP txs' non-token parents
	// resolve as definitively not SLP (no lokad at vout 0) rather than
	// unknown
	_, fundUtxo, err := s.fund(1e6)
	if err != nil {
		return err
	}

	// Token lifecycle: genesis -> send -> mint, all valid
	genesisTx, err := build.TokenCreate(build.TokenCreateRequest{
		Wallet:   s.wallet(fundUtxo),
		SlpType:  memo.SlpDefaultTokenType,
		Ticker:   "TEST",
		Name:     "Test Token",
		Quantity: 1000,
	})
	if err != nil {
		return fmt.Errorf("error building genesis; %w", err)
	}
	tokenHash := genesisTx.GetHash()
	if err := s.save(genesisTx); err != nil {
		return err
	}
	if err := s.checkStatus("genesis", genesisTx, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}

	send1, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(tokenUtxoAt(genesisTx, 1, tokenHash, 1000), bchChange(genesisTx)),
		TokenHash: tokenHash,
		Recipient: test_tx.Address2,
		Quantity:  400,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building send1; %w", err)
	}
	if err := s.save(send1); err != nil {
		return err
	}
	if err := s.checkStatus("send with valid inputs", send1, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}

	mintTx, err := build.TokenMint(build.TokenMintRequest{
		Wallet:       s.wallet(bchChange(send1)),
		Baton:        tokenUtxoAt(genesisTx, 2, tokenHash, 0),
		BatonAddress: test_tx.Address1,
		TokenAddress: test_tx.Address1,
		TokenHash:    tokenHash,
		TokenType:    memo.SlpDefaultTokenType,
		Quantity:     500,
	})
	if err != nil {
		return fmt.Errorf("error building mint; %w", err)
	}
	if err := s.save(mintTx); err != nil {
		return err
	}
	if err := s.checkStatus("mint with valid baton", mintTx, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}

	// Fake transfer: declares token outputs but its input is a plain BCH
	// output the wallet lies about; the index must call it invalid
	_, fakeFundUtxo, err := s.fund(8e5)
	if err != nil {
		return err
	}
	fakeFundUtxo.SlpToken = tokenHash
	fakeFundUtxo.SlpQuantity = 1000
	fakeSend, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(fakeFundUtxo),
		TokenHash: tokenHash,
		Recipient: test_tx.Address2,
		Quantity:  700,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building fake send; %w", err)
	}
	if err := s.save(fakeSend); err != nil {
		return err
	}
	if err := s.checkStatus("fake send (no valid inputs)", fakeSend,
		slp_tx.StatusInvalid, slp_tx.ReasonSendInputSum); err != nil {
		return err
	}

	// Fake mint: "baton" input is a plain BCH output
	_, fakeBatonUtxo, err := s.fund(8e5)
	if err != nil {
		return err
	}
	fakeBatonUtxo.SlpToken = tokenHash
	fakeMint, err := build.TokenMint(build.TokenMintRequest{
		Wallet:       s.wallet(),
		Baton:        fakeBatonUtxo,
		BatonAddress: test_tx.Address1,
		TokenAddress: test_tx.Address1,
		TokenHash:    tokenHash,
		TokenType:    memo.SlpDefaultTokenType,
		Quantity:     999999,
	})
	if err != nil {
		return fmt.Errorf("error building fake mint; %w", err)
	}
	if err := s.save(fakeMint); err != nil {
		return err
	}
	if err := s.checkStatus("fake mint (no baton)", fakeMint,
		slp_tx.StatusInvalid, slp_tx.ReasonMintNoBaton); err != nil {
		return err
	}

	// Invalid verdicts cascade too: two generations of descendants of a
	// phantom send are saved first (pending), then the ancestor's invalid
	// verdict must propagate through both, with the input-sum reason (an
	// invalid parent's declared outputs contribute zero, they don't taint)
	_, phantomFundUtxo, err := s.fund(8e5)
	if err != nil {
		return err
	}
	phantomFundUtxo.SlpToken = tokenHash
	phantomFundUtxo.SlpQuantity = 1000
	phantomParent, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(phantomFundUtxo),
		TokenHash: tokenHash,
		Recipient: test_tx.Address2,
		Quantity:  700,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building phantom parent; %w", err)
	}
	phantomChange, err := tokenChange(phantomParent, tokenHash, 300)
	if err != nil {
		return err
	}
	phantomChild, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(phantomChange, bchChange(phantomParent)),
		TokenHash: tokenHash,
		Recipient: test_tx.Address2,
		Quantity:  100,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building phantom child; %w", err)
	}
	phantomChildChange, err := tokenChange(phantomChild, tokenHash, 200)
	if err != nil {
		return err
	}
	phantomGrandchild, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(phantomChildChange, bchChange(phantomChild)),
		TokenHash: tokenHash,
		Recipient: test_tx.Address2,
		Quantity:  150,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building phantom grandchild; %w", err)
	}
	if err := s.save(phantomGrandchild); err != nil {
		return err
	}
	if err := s.checkStatus("phantom grandchild before ancestors", phantomGrandchild,
		slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
		return err
	}
	if err := s.save(phantomChild); err != nil {
		return err
	}
	if err := s.checkStatus("phantom child before parent", phantomChild,
		slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
		return err
	}
	if err := s.save(phantomParent); err != nil {
		return err
	}
	if err := s.checkStatus("phantom parent", phantomParent,
		slp_tx.StatusInvalid, slp_tx.ReasonSendInputSum); err != nil {
		return err
	}
	if err := s.checkStatus("phantom child invalid by cascade", phantomChild,
		slp_tx.StatusInvalid, slp_tx.ReasonSendInputSum); err != nil {
		return err
	}
	if err := s.checkStatus("phantom grandchild invalid by cascade", phantomGrandchild,
		slp_tx.StatusInvalid, slp_tx.ReasonSendInputSum); err != nil {
		return err
	}

	// Out-of-order arrival: child saved before its parent exists anywhere.
	// Must be pending (never invalid), then valid once the parent is known
	// and the child is revisited (the confirmation-pass retry).
	send1Change, err := tokenChange(send1, tokenHash, 600)
	if err != nil {
		return err
	}
	sendParent, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(send1Change, bchChange(mintTx)),
		TokenHash: tokenHash,
		Recipient: test_tx.Address2,
		Quantity:  100,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building parent send; %w", err)
	}
	parentChange, err := tokenChange(sendParent, tokenHash, 500)
	if err != nil {
		return err
	}
	sendChild, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(parentChange, bchChange(sendParent)),
		TokenHash: tokenHash,
		Recipient: test_tx.Address2,
		Quantity:  250,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building child send; %w", err)
	}
	if err := s.save(sendChild); err != nil {
		return err
	}
	if err := s.checkStatus("child before parent", sendChild, slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
		return err
	}
	if err := s.save(sendParent); err != nil {
		return err
	}
	if err := s.checkStatus("parent arrives", sendParent, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	if err := s.checkStatus("child resolved by cascade", sendChild, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}

	// Deep chain: descendants A -> B -> C -> D saved deepest-first, so three
	// generations are pending at once. When the ancestor arrives, cascading
	// validation must walk the whole spend graph in one pass.
	sendChildChange, err := tokenChange(sendChild, tokenHash, 250)
	if err != nil {
		return err
	}
	var chainTxs []*memo.Tx
	var prevToken = sendChildChange
	var prevBch = bchChange(sendChild)
	var prevQuantity = uint64(250)
	for i := 0; i < 4; i++ {
		chainTx, err := build.TokenSend(build.TokenSendRequest{
			Wallet:    s.wallet(prevToken, prevBch),
			TokenHash: tokenHash,
			Recipient: test_tx.Address2,
			Quantity:  10,
			TokenType: memo.SlpDefaultTokenType,
		})
		if err != nil {
			return fmt.Errorf("error building chain tx %d; %w", i, err)
		}
		chainTxs = append(chainTxs, chainTx)
		prevQuantity -= 10
		prevToken, err = tokenChange(chainTx, tokenHash, prevQuantity)
		if err != nil {
			return err
		}
		prevBch = bchChange(chainTx)
	}
	// Save D, C, B (parents unseen: all pending), then ancestor A (valid)
	for i := 3; i >= 1; i-- {
		if err := s.save(chainTxs[i]); err != nil {
			return err
		}
		if err := s.checkStatus(fmt.Sprintf("chain generation %d before ancestors", i),
			chainTxs[i], slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
			return err
		}
	}
	if err := s.save(chainTxs[0]); err != nil {
		return err
	}
	if err := s.checkStatus("chain ancestor", chainTxs[0], slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	for i := 1; i < 4; i++ {
		if err := s.checkStatus(fmt.Sprintf("chain generation %d after cascade", i),
			chainTxs[i], slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
			return err
		}
	}

	// Index sweep backfill: wipe the verdicts of the whole out-of-order graph
	// to simulate history ingested before validation existed, and let one
	// index sweep (driven by the slp topics, txhash order) re-derive them all
	var backfillTxs = append([]*memo.Tx{sendChild, sendParent}, chainTxs...)
	var removeObjects []db.Object
	for _, tx := range backfillTxs {
		removeObjects = append(removeObjects, &item_slp.Validity{TxHash: txHash32(tx)})
	}
	if err := db.Remove(removeObjects); err != nil {
		return fmt.Errorf("error removing validities for backfill simulation; %w", err)
	}
	for _, tx := range backfillTxs {
		if err := s.checkStatus("backfill wiped verdict", tx, slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
			return err
		}
	}
	if err := act_maint.NewSlpValiditySweep(s.ctx, false).Run(); err != nil {
		return fmt.Errorf("error running slp validity index sweep; %w", err)
	}
	for _, tx := range backfillTxs {
		if err := s.checkStatus("backfill re-derived verdict", tx, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
			return err
		}
	}
	log.Printf("✓ index sweep re-derived wiped verdicts\n")

	// Deep backfill: strip a tx's verdict and transcription rows so the
	// index sweep cannot enumerate it (simulating a tx the live path never
	// transcribed), then verify only the chain-tx-output backfill finds it.
	// The backfill skips unmined txs, so the gap tx is mined first.
	var gapTx = chainTxs[3] // last generation: nothing decided cascades to it
	var gapHash = txHash32(gapTx)
	if err := db.Remove([]db.Object{
		&item_slp.Validity{TxHash: gapHash},
		&item_slp.Send{TxHash: gapHash},
		&item_slp.Output{TxHash: gapHash, Index: 1},
		&item_slp.Output{TxHash: gapHash, Index: 2},
	}); err != nil {
		return fmt.Errorf("error removing gap tx rows for backfill simulation; %w", err)
	}
	if err := act_maint.NewSlpValiditySweep(s.ctx, false).Run(); err != nil {
		return fmt.Errorf("error running slp validity index sweep with gap tx; %w", err)
	}
	if err := s.checkStatus("gap tx invisible to index sweep", gapTx, slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
		return err
	}
	if err := s.mine(700011, 11, gapTx); err != nil {
		return err
	}
	if err := act_maint.NewSlpValidityBackfill(s.ctx, false).Run(); err != nil {
		return fmt.Errorf("error running slp validity backfill for gap tx; %w", err)
	}
	if err := s.checkStatus("gap tx found by backfill", gapTx, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	sendHashes, err := item_slp.GetTopicTxHashes(s.ctx, db.TopicSlpSend, db.GetShardIdFromByte32(gapHash[:]), nil)
	if err != nil {
		return fmt.Errorf("error getting slp send hashes after backfill; %w", err)
	}
	var sendRestored bool
	for _, sendHash := range sendHashes {
		if sendHash == gapHash {
			sendRestored = true
			break
		}
	}
	if !sendRestored {
		return fmt.Errorf("error gap tx slp send row not restored by backfill")
	}
	log.Printf("✓ backfill transcribed and validated the gap tx\n")

	// Repopulation window: the wipe-and-repopulate deploy leaves chain rows
	// intact while slp rows are missing until the async repopulation reaches
	// them. Simulate a send whose token's genesis is not yet repopulated: the
	// genesis's transcription rows and the pair's verdicts are stripped, and
	// the index sweep must leave the send pending - never invalid - until the
	// genesis is re-transcribed (backfill), whose verdict then resolves it.
	var genesisHash = txHash32(genesisTx)
	if err := db.Remove([]db.Object{
		&item_slp.Genesis{TxHash: genesisHash},
		&item_slp.Output{TxHash: genesisHash, Index: 1},
		&item_slp.Baton{TxHash: genesisHash, Index: 2},
		&item_slp.Validity{TxHash: genesisHash},
		&item_slp.Validity{TxHash: txHash32(send1)},
	}); err != nil {
		return fmt.Errorf("error removing genesis rows for repopulation simulation; %w", err)
	}
	if err := act_maint.NewSlpValiditySweep(s.ctx, false).Run(); err != nil {
		return fmt.Errorf("error running slp validity index sweep with missing genesis; %w", err)
	}
	if err := s.checkStatus("send with unrepopulated genesis stays pending", send1,
		slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
		return err
	}
	// One block holding parent and child: the in-chunk topo sort must put the
	// genesis first and the send resolves in the same validation call
	if err := s.mine(700012, 12, genesisTx, send1); err != nil {
		return err
	}
	if err := act_maint.NewSlpValidityBackfill(s.ctx, false).Run(); err != nil {
		return fmt.Errorf("error running slp validity backfill to repopulate genesis; %w", err)
	}
	if err := s.checkStatus("genesis repopulated by backfill", genesisTx,
		slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	if err := s.checkStatus("send resolved after genesis repopulated", send1,
		slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	log.Printf("✓ missing genesis left send pending until repopulated\n")

	// Cascade transcription: the cascade writes verdicts for spenders the
	// sweep never enumerated (their slp rows are missing, e.g. mid-backfill),
	// and it must transcribe them as it decides them. A decided parent's
	// missing output rows read as definitively contributing nothing, so an
	// untranscribed middle generation would falsely invalidate every deeper
	// descendant. Wipe the whole chain's verdicts plus generations 1-3's
	// transcription rows, leaving only the ancestor enumerable by the index
	// sweep; its verdict must cascade down with rows restored at every step.
	var cascadeWipe = []db.Object{&item_slp.Validity{TxHash: txHash32(chainTxs[0])}}
	for i := 1; i < 4; i++ {
		chainHash := txHash32(chainTxs[i])
		cascadeWipe = append(cascadeWipe,
			&item_slp.Validity{TxHash: chainHash},
			&item_slp.Send{TxHash: chainHash},
			&item_slp.Output{TxHash: chainHash, Index: 1},
			&item_slp.Output{TxHash: chainHash, Index: 2},
		)
	}
	if err := db.Remove(cascadeWipe); err != nil {
		return fmt.Errorf("error removing chain rows for cascade transcription phase; %w", err)
	}
	for i := 0; i < 4; i++ {
		if err := s.checkStatus(fmt.Sprintf("chain generation %d wiped for cascade transcription", i),
			chainTxs[i], slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
			return err
		}
	}
	if err := act_maint.NewSlpValiditySweep(s.ctx, false).Run(); err != nil {
		return fmt.Errorf("error running slp validity sweep for cascade transcription phase; %w", err)
	}
	for i := 0; i < 4; i++ {
		if err := s.checkStatus(fmt.Sprintf("chain generation %d after sweep-driven cascade", i),
			chainTxs[i], slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
			return err
		}
	}
	for i := 1; i < 4; i++ {
		chainHash := txHash32(chainTxs[i])
		outputs, err := item_slp.GetOutputs(s.ctx, []memo.Out{{TxHash: chainHash[:], Index: 1}})
		if err != nil {
			return fmt.Errorf("error getting chain generation %d output row; %w", i, err)
		}
		if len(outputs) == 0 {
			return fmt.Errorf("error chain generation %d output row not restored by cascade", i)
		}
	}
	log.Printf("✓ sweep-driven cascade transcribed and validated unenumerated spenders\n")

	// Once-per-cascade transcription: a fan-in spender re-enters the cascade
	// frontier once per round in which another of its parents decides, and
	// must be re-validated each time (a new parent verdict can tip it) but
	// transcribed at most once per cascade — per-round re-transcription is the
	// quadratic write churn that stalled the production audit. Graph:
	// gen2 -> f1 -> f2, with fanIn spending outputs of both f1 and f2. Seeded
	// from gen2 alone, fanIn enters the frontier as f1's spender (parks
	// pending, f2 undecided that round) and again as f2's spender (decides).
	_, fanFundUtxo, err := s.fund(1e6)
	if err != nil {
		return err
	}
	gen2Tx, err := build.TokenCreate(build.TokenCreateRequest{
		Wallet:   s.wallet(fanFundUtxo),
		SlpType:  memo.SlpDefaultTokenType,
		Ticker:   "FANIN",
		Name:     "Fan In Token",
		Quantity: 100,
	})
	if err != nil {
		return fmt.Errorf("error building fan-in genesis; %w", err)
	}
	fanHash := gen2Tx.GetHash()
	if err := s.save(gen2Tx); err != nil {
		return err
	}
	f1, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(tokenUtxoAt(gen2Tx, 1, fanHash, 100), bchChange(gen2Tx)),
		TokenHash: fanHash,
		Recipient: test_tx.Address1,
		Quantity:  40,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building fan-in f1; %w", err)
	}
	f2, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(tokenUtxoAt(f1, 2, fanHash, 60), bchChange(f1)),
		TokenHash: fanHash,
		Recipient: test_tx.Address1,
		Quantity:  15,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building fan-in f2; %w", err)
	}
	fanIn, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(tokenUtxoAt(f1, 1, fanHash, 40), tokenUtxoAt(f2, 1, fanHash, 15), bchChange(f2)),
		TokenHash: fanHash,
		Recipient: test_tx.Address1,
		Quantity:  55,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building fan-in spender; %w", err)
	}
	// Saved parents-first: each tx must be decidable at save time
	for _, fanTx := range []struct {
		name string
		tx   *memo.Tx
	}{{"f1", f1}, {"f2", f2}, {"fan-in spender", fanIn}} {
		if err := s.save(fanTx.tx); err != nil {
			return err
		}
		if err := s.checkStatus("fan-in "+fanTx.name+" live", fanTx.tx, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
			return err
		}
	}
	// Strip the graph back to only gen2 being enumerable, undecided everywhere
	var fanWipe = []db.Object{&item_slp.Validity{TxHash: txHash32(gen2Tx)}}
	for _, tx := range []*memo.Tx{f1, f2, fanIn} {
		fanTxHash := txHash32(tx)
		fanWipe = append(fanWipe,
			&item_slp.Validity{TxHash: fanTxHash},
			&item_slp.Send{TxHash: fanTxHash},
			&item_slp.Output{TxHash: fanTxHash, Index: 1},
			&item_slp.Output{TxHash: fanTxHash, Index: 2},
		)
	}
	if err := db.Remove(fanWipe); err != nil {
		return fmt.Errorf("error removing fan-in rows; %w", err)
	}
	var transcribeCounts = make(map[[32]byte]int)
	origCascadeTranscribe := slp_validate.CascadeTranscribe
	slp_validate.CascadeTranscribe = func(txs []slp_validate.Tx) error {
		for _, tx := range txs {
			transcribeCounts[tx.TxHash]++
		}
		return origCascadeTranscribe(txs)
	}
	sweepErr := act_maint.NewSlpValiditySweep(s.ctx, false).Run()
	if sweepErr != nil {
		slp_validate.CascadeTranscribe = origCascadeTranscribe
		return fmt.Errorf("error running sweep for fan-in phase; %w", sweepErr)
	}
	for _, fanTx := range []struct {
		name string
		tx   *memo.Tx
	}{{"gen2", gen2Tx}, {"f1", f1}, {"f2", f2}, {"fan-in spender", fanIn}} {
		if err := s.checkStatus("fan-in "+fanTx.name+" after cascade", fanTx.tx, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
			slp_validate.CascadeTranscribe = origCascadeTranscribe
			return err
		}
		if fanTx.tx == gen2Tx {
			continue // gen2 is the sweep's own candidate, not a cascade spender
		}
		if got := transcribeCounts[txHash32(fanTx.tx)]; got != 1 {
			slp_validate.CascadeTranscribe = origCascadeTranscribe
			return fmt.Errorf("error fan-in %s transcribed %d times in one cascade, expected exactly 1", fanTx.name, got)
		}
	}
	// Per-cascade lifetime: a NEW cascade must transcribe a wiped spender
	// again — the memo is per invocation, not global
	if err := db.Remove([]db.Object{
		&item_slp.Validity{TxHash: txHash32(f2)},
		&item_slp.Validity{TxHash: txHash32(fanIn)},
		&item_slp.Send{TxHash: txHash32(fanIn)},
		&item_slp.Output{TxHash: txHash32(fanIn), Index: 1},
	}); err != nil {
		slp_validate.CascadeTranscribe = origCascadeTranscribe
		return fmt.Errorf("error removing fan-in spender rows for second cascade; %w", err)
	}
	sweepErr = act_maint.NewSlpValiditySweep(s.ctx, false).Run()
	slp_validate.CascadeTranscribe = origCascadeTranscribe
	if sweepErr != nil {
		return fmt.Errorf("error running second sweep for fan-in phase; %w", sweepErr)
	}
	if err := s.checkStatus("fan-in spender after second cascade", fanIn, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	if got := transcribeCounts[txHash32(fanIn)]; got != 2 {
		return fmt.Errorf("error fan-in spender transcribed %d times across two cascades, expected exactly 2", got)
	}
	if got := transcribeCounts[txHash32(f1)]; got != 1 {
		return fmt.Errorf("error fan-in f1 transcribed %d times, expected exactly 1 (decided txs must not re-transcribe)", got)
	}
	log.Printf("✓ fan-in spender transcribed once per cascade, revalidated to valid\n")

	// Secondary-message poison: a tx with a valid token-A SEND at vout 0 and
	// a second OP_RETURN carrying a SEND of a different token B at a later
	// output. Per spec (Consideration A) only vout 0 is an SLP action: the
	// vout-1 message must not be transcribed (no token-B rows, no overwrite of
	// the vout-0 Send row) and must not affect the verdict. Before the vout-0
	// gate, the later message's rows (keyed by txhash) overwrote the vout-0
	// rows, enabling a false-valid child spend of a real token B.
	_, poisonBchUtxo, err := s.fund(8e5)
	if err != nil {
		return err
	}
	poisonTx, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(tokenUtxoAt(mintTx, 1, tokenHash, 500), poisonBchUtxo),
		TokenHash: tokenHash,
		Recipient: test_tx.Address2,
		Quantity:  200,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building poison tx; %w", err)
	}
	var tokenB [32]byte
	for i := range tokenB {
		tokenB[i] = 0xbb
	}
	secondaryScript, err := script.TokenSend{
		TokenHash:  tokenB[:],
		SlpType:    memo.SlpDefaultTokenType,
		Quantities: []uint64{0, 0, 999999}, // assign token B to a later vout
	}.Get()
	if err != nil {
		return fmt.Errorf("error building secondary slp script; %w", err)
	}
	poisonTx.MsgTx.AddTxOut(wire.NewTxOut(0, secondaryScript))
	if err := s.save(poisonTx); err != nil {
		return err
	}
	if err := s.checkStatus("poison tx (valid token-A send)", poisonTx,
		slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	var poisonSend = &item_slp.Send{TxHash: txHash32(poisonTx)}
	if err := db.GetItem(s.ctx, poisonSend); err != nil {
		return fmt.Errorf("error getting poison tx send row; %w", err)
	}
	var tokenAHash [32]byte
	copy(tokenAHash[:], tokenHash)
	if poisonSend.TokenHash != tokenAHash {
		return fmt.Errorf("error poison tx send row token hash %x, expected token A %x (secondary message overwrote it)",
			poisonSend.TokenHash, tokenHash)
	}
	poisonOuts, err := item_slp.GetOutputs(s.ctx, []memo.Out{
		{TxHash: poisonTx.GetHash(), Index: 1},
		{TxHash: poisonTx.GetHash(), Index: 2},
		{TxHash: poisonTx.GetHash(), Index: 3},
	})
	if err != nil {
		return fmt.Errorf("error getting poison tx outputs; %w", err)
	}
	for _, out := range poisonOuts {
		if out.TokenHash == tokenB {
			return fmt.Errorf("error poison tx output %d carries token B row (secondary message was transcribed)", out.Index)
		}
	}
	log.Printf("✓ secondary op_return message left no token-B rows\n")

	// Backfill re-poison guard: strip the poison tx's derived rows + verdict,
	// mine it, and run the backfill. The tx carries a token-A message at vout
	// 0 and a token-B message at a later output; the backfill must
	// re-transcribe only the vout-0 message (the transcribe() vout-0 gate),
	// restoring token-A rows and never re-poisoning token-B rows. This
	// exercises the recovery path, which the live-handler poison assertion
	// above does not.
	poisonHash := txHash32(poisonTx)
	if err := db.Remove([]db.Object{
		&item_slp.Send{TxHash: poisonHash},
		&item_slp.Output{TxHash: poisonHash, Index: 1},
		&item_slp.Output{TxHash: poisonHash, Index: 2},
		&item_slp.Validity{TxHash: poisonHash},
	}); err != nil {
		return fmt.Errorf("error removing poison tx rows for backfill guard; %w", err)
	}
	if err := s.mine(700013, 13, poisonTx); err != nil {
		return err
	}
	if err := act_maint.NewSlpValidityBackfill(s.ctx, false).Run(); err != nil {
		return fmt.Errorf("error running backfill for poison guard; %w", err)
	}
	if err := s.checkStatus("poison tx re-validated by backfill", poisonTx,
		slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	var poisonSendBackfill = &item_slp.Send{TxHash: poisonHash}
	if err := db.GetItem(s.ctx, poisonSendBackfill); err != nil {
		return fmt.Errorf("error getting poison tx send row after backfill; %w", err)
	}
	if poisonSendBackfill.TokenHash != tokenAHash {
		return fmt.Errorf("error backfill restored poison send row as token %x, expected token A (backfill re-transcribed a later message)",
			poisonSendBackfill.TokenHash)
	}
	poisonOutsBackfill, err := item_slp.GetOutputs(s.ctx, []memo.Out{
		{TxHash: poisonTx.GetHash(), Index: 1},
		{TxHash: poisonTx.GetHash(), Index: 2},
		{TxHash: poisonTx.GetHash(), Index: 3},
	})
	if err != nil {
		return fmt.Errorf("error getting poison tx outputs after backfill; %w", err)
	}
	var tokenAOuts int
	for _, out := range poisonOutsBackfill {
		if out.TokenHash == tokenB {
			return fmt.Errorf("error backfill re-poisoned output %d with token B", out.Index)
		}
		if out.TokenHash == tokenAHash {
			tokenAOuts++
		}
	}
	if tokenAOuts != 2 {
		return fmt.Errorf("error backfill restored %d token-A outputs, expected 2", tokenAOuts)
	}
	log.Printf("✓ backfill re-transcribed only the vout-0 message\n")

	// Cascade later-lokad spender guard: a tx with a plain (non-SLP) vout 0
	// and an SLP lokad only in a later output is not an SLP action. When it is
	// discovered as a spender during cascading validation, ReconstructSlpTxs's
	// vout-0 check must drop it: no verdict, no rows. Without that check it
	// would reconstruct off the later lokad and receive a spurious verdict.
	// Built through the cascade (spender saved first, parent second) rather
	// than the live handler, so it exercises ReconstructSlpTxs specifically.
	_, cascadeBchUtxo, err := s.fund(8e5)
	if err != nil {
		return err
	}
	cascadeParent, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(tokenUtxoAt(poisonTx, 2, tokenHash, 300), cascadeBchUtxo),
		TokenHash: tokenHash,
		Recipient: test_tx.Address2,
		Quantity:  100,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building cascade parent; %w", err)
	}
	laterLokadScript, err := script.TokenSend{
		TokenHash:  tokenB[:],
		SlpType:    memo.SlpDefaultTokenType,
		Quantities: []uint64{999999},
	}.Get()
	if err != nil {
		return fmt.Errorf("error building later lokad script; %w", err)
	}
	cascadeParentHash := chainhash.Hash(txHash32(cascadeParent))
	changeIndex := uint32(len(cascadeParent.MsgTx.TxOut) - 1)
	spenderMsgTx := wire.NewMsgTx(1)
	spenderMsgTx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(&cascadeParentHash, changeIndex), nil))
	// vout 0 is a plain (non-OP_RETURN) output reusing the parent's change
	// lock; the SLP lokad lives only at vout 1
	spenderMsgTx.AddTxOut(wire.NewTxOut(1000, cascadeParent.MsgTx.TxOut[changeIndex].PkScript))
	spenderMsgTx.AddTxOut(wire.NewTxOut(0, laterLokadScript))
	cascadeSpender := &memo.Tx{MsgTx: spenderMsgTx}
	if err := s.save(cascadeSpender); err != nil {
		return err
	}
	if err := s.checkStatus("cascade spender before parent (ignored)", cascadeSpender,
		slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
		return err
	}
	if err := s.save(cascadeParent); err != nil {
		return err
	}
	if err := s.checkStatus("cascade parent valid", cascadeParent,
		slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	if err := s.checkStatus("cascade later-lokad spender still ignored", cascadeSpender,
		slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
		return err
	}
	// Absence must be the not-found sentinel specifically; any other error
	// means the lookup itself failed and absence was never established
	if err := db.GetItem(s.ctx, &item_slp.Send{TxHash: txHash32(cascadeSpender)}); err == nil {
		return fmt.Errorf("error cascade spender has an slp send row (later message was transcribed)")
	} else if !client.IsEntryNotFoundError(err) {
		return fmt.Errorf("error checking cascade spender send row absence; %w", err)
	}
	log.Printf("✓ cascade dropped later-lokad spender (no verdict, no rows)\n")

	// NFT1: group genesis, child genesis spending the group token at input
	// 0, and a fake child whose input 0 is a plain BCH output
	_, nftFundUtxo, err := s.fund(9e5)
	if err != nil {
		return err
	}
	groupTx, err := build.TokenCreate(build.TokenCreateRequest{
		Wallet:   s.wallet(nftFundUtxo),
		SlpType:  memo.SlpNftGroupTokenType,
		Ticker:   "GRP",
		Name:     "Test Group",
		Quantity: 10,
	})
	if err != nil {
		return fmt.Errorf("error building nft group genesis; %w", err)
	}
	if err := s.save(groupTx); err != nil {
		return err
	}
	if err := s.checkStatus("nft group genesis", groupTx, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	groupUtxo := tokenUtxoAt(groupTx, 1, groupTx.GetHash(), 10)
	childTx, err := build.TokenCreate(build.TokenCreateRequest{
		Wallet:   s.wallet(bchChange(groupTx)),
		SlpType:  memo.SlpNftChildTokenType,
		Ticker:   "KID",
		Name:     "Test Child",
		Quantity: 1,
		NftUtxo:  &groupUtxo,
	})
	if err != nil {
		return fmt.Errorf("error building nft child genesis; %w", err)
	}
	if err := s.save(childTx); err != nil {
		return err
	}
	if err := s.checkStatus("nft child genesis", childTx, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	_, fakeGroupUtxo, err := s.fund(9e5)
	if err != nil {
		return err
	}
	fakeGroupUtxo.SlpToken = groupTx.GetHash()
	fakeGroupUtxo.SlpQuantity = 1
	fakeChildTx, err := build.TokenCreate(build.TokenCreateRequest{
		Wallet:   s.wallet(),
		SlpType:  memo.SlpNftChildTokenType,
		Ticker:   "BAD",
		Name:     "Fake Child",
		Quantity: 1,
		NftUtxo:  &fakeGroupUtxo,
	})
	if err != nil {
		return fmt.Errorf("error building fake nft child genesis; %w", err)
	}
	if err := s.save(fakeChildTx); err != nil {
		return err
	}
	if err := s.checkStatus("fake nft child (no group input)", fakeChildTx,
		slp_tx.StatusInvalid, slp_tx.ReasonNftChildGenesis); err != nil {
		return err
	}

	// Address-less inputs: SLP dispatch must not depend on resolving a sender
	// address from the unlock scripts (memo handlers skip such txs; the old
	// saver skipped SLP txs too). Blank the send's sig scripts so no input
	// parses as P2PKH/P2SH; the send must still be transcribed and validated
	// against its parents. Blanking changes the txid, which is fine — the
	// prevouts linking it to its parents are untouched and sigs are never
	// verified by the pipeline.
	_, noAddrFundUtxo, err := s.fund(7e5)
	if err != nil {
		return err
	}
	noAddrGenesisTx, err := build.TokenCreate(build.TokenCreateRequest{
		Wallet:   s.wallet(noAddrFundUtxo),
		SlpType:  memo.SlpDefaultTokenType,
		Ticker:   "NOAD",
		Name:     "No Addr Token",
		Quantity: 25,
	})
	if err != nil {
		return fmt.Errorf("error building no-addr genesis; %w", err)
	}
	noAddrTokenHash := noAddrGenesisTx.GetHash()
	if err := s.save(noAddrGenesisTx); err != nil {
		return err
	}
	noAddrSendTx, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(tokenUtxoAt(noAddrGenesisTx, 1, noAddrTokenHash, 25), bchChange(noAddrGenesisTx)),
		TokenHash: noAddrTokenHash,
		Recipient: test_tx.Address2,
		Quantity:  10,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building no-addr send; %w", err)
	}
	for _, txIn := range noAddrSendTx.MsgTx.TxIn {
		txIn.SignatureScript = nil
	}
	if err := s.save(noAddrSendTx); err != nil {
		return err
	}
	if err := s.checkStatus("send with address-less inputs", noAddrSendTx,
		slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	var noAddrSend = &item_slp.Send{TxHash: txHash32(noAddrSendTx)}
	if err := db.GetItem(s.ctx, noAddrSend); err != nil {
		return fmt.Errorf("error getting no-addr send row (tx not transcribed); %w", err)
	}

	// Block-ordered backfill: a fresh token graph (genesis -> three chained
	// sends) is mined across two blocks, the sends deliberately in reverse
	// in-block order (CTOR orders by txid, not topologically), and every
	// verdict and transcription row wiped. slp-validity-backfill must
	// re-derive the whole graph in a single validation pass per chunk: the
	// topo sort puts parents first and in-call verdict visibility resolves
	// the chain without cascade rounds, so nothing may end up pending and
	// nothing may be skipped as unmined.
	_, bfFundUtxo, err := s.fund(1e6)
	if err != nil {
		return err
	}
	bfGenesisTx, err := build.TokenCreate(build.TokenCreateRequest{
		Wallet:   s.wallet(bfFundUtxo),
		SlpType:  memo.SlpDefaultTokenType,
		Ticker:   "BKFL",
		Name:     "Backfill Token",
		Quantity: 90,
	})
	if err != nil {
		return fmt.Errorf("error building backfill genesis; %w", err)
	}
	bfTokenHash := bfGenesisTx.GetHash()
	if err := s.save(bfGenesisTx); err != nil {
		return err
	}
	var bfSends []*memo.Tx
	var bfPrevToken = tokenUtxoAt(bfGenesisTx, 1, bfTokenHash, 90)
	var bfPrevBch = bchChange(bfGenesisTx)
	var bfPrevQuantity = uint64(90)
	for i := 0; i < 3; i++ {
		bfSend, err := build.TokenSend(build.TokenSendRequest{
			Wallet:    s.wallet(bfPrevToken, bfPrevBch),
			TokenHash: bfTokenHash,
			Recipient: test_tx.Address2,
			Quantity:  10,
			TokenType: memo.SlpDefaultTokenType,
		})
		if err != nil {
			return fmt.Errorf("error building backfill send %d; %w", i, err)
		}
		if err := s.save(bfSend); err != nil {
			return err
		}
		bfSends = append(bfSends, bfSend)
		bfPrevQuantity -= 10
		bfPrevToken, err = tokenChange(bfSend, bfTokenHash, bfPrevQuantity)
		if err != nil {
			return err
		}
		bfPrevBch = bchChange(bfSend)
	}
	var bfTxs = append([]*memo.Tx{bfGenesisTx}, bfSends...)
	bfGenesisHash := txHash32(bfGenesisTx)
	var bfWipe = []db.Object{
		&item_slp.Validity{TxHash: bfGenesisHash},
		&item_slp.Genesis{TxHash: bfGenesisHash},
		&item_slp.Output{TxHash: bfGenesisHash, Index: 1},
		&item_slp.Baton{TxHash: bfGenesisHash, Index: 2},
	}
	for _, bfSend := range bfSends {
		bfSendHash := txHash32(bfSend)
		bfWipe = append(bfWipe,
			&item_slp.Validity{TxHash: bfSendHash},
			&item_slp.Send{TxHash: bfSendHash},
			&item_slp.Output{TxHash: bfSendHash, Index: 1},
			&item_slp.Output{TxHash: bfSendHash, Index: 2},
		)
	}
	if err := db.Remove(bfWipe); err != nil {
		return fmt.Errorf("error removing backfill graph rows; %w", err)
	}
	for _, bfTx := range bfTxs {
		if err := s.checkStatus("backfill graph wiped", bfTx, slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
			return err
		}
	}
	// Mine: genesis alone at the lower height; the sends in one later block
	// in child-first in-block order (mine skips the live SLP handler, which
	// would re-validate here)
	if err := s.mine(700001, 1, bfGenesisTx); err != nil {
		return err
	}
	if err := s.mine(700002, 2, bfSends[2], bfSends[1], bfSends[0]); err != nil {
		return err
	}
	backfill := act_maint.NewSlpValidityBackfill(s.ctx, false)
	if err := backfill.Run(); err != nil {
		return fmt.Errorf("error running slp validity backfill; %w", err)
	}
	for _, bfTx := range bfTxs {
		if err := s.checkStatus("backfill re-derived", bfTx, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
			return err
		}
	}
	// Single-pass proof: nothing parked pending (the CTOR-reversed chain
	// resolved inside its chunk) and nothing was skipped as unmined
	if backfill.Pending != 0 {
		return fmt.Errorf("error backfill left %d txs pending, expected single-pass resolution", backfill.Pending)
	}
	if backfill.MempoolTail != 0 {
		return fmt.Errorf("error backfill skipped %d txs as unmined, expected all mined", backfill.MempoolTail)
	}
	if backfill.SlpTxs != int64(len(bfTxs)) {
		return fmt.Errorf("error backfill validated %d txs, expected exactly the %d wiped ones", backfill.SlpTxs, len(bfTxs))
	}
	if backfill.Missing != 0 {
		return fmt.Errorf("error backfill reported %d incomplete txs", backfill.Missing)
	}
	// Transcription rows must be restored by the backfill's batched pass
	var bfGenesisRow = &item_slp.Genesis{TxHash: bfGenesisHash}
	if err := db.GetItem(s.ctx, bfGenesisRow); err != nil {
		return fmt.Errorf("error backfill genesis row not restored; %w", err)
	}
	for i, bfSend := range bfSends {
		bfSendHash := txHash32(bfSend)
		bfOutputs, err := item_slp.GetOutputs(s.ctx, []memo.Out{{TxHash: bfSendHash[:], Index: 1}})
		if err != nil {
			return fmt.Errorf("error getting backfill send %d output row; %w", i, err)
		}
		if len(bfOutputs) == 0 {
			return fmt.Errorf("error backfill send %d output row not restored", i)
		}
	}
	log.Printf("✓ block-ordered backfill re-derived the CTOR-reversed graph in one pass\n")

	// Unmined skip: an SLP candidate with chain rows but no mined block must
	// be counted and left untouched by the backfill — no transcription, no
	// verdict — then picked up by a re-run once mined. Saved via the minimal
	// saver (headerless = mempool) so the live SLP handler never sees it.
	_, umFundUtxo, err := s.fund(1e6)
	if err != nil {
		return err
	}
	umGenesisTx, err := build.TokenCreate(build.TokenCreateRequest{
		Wallet:   s.wallet(umFundUtxo),
		SlpType:  memo.SlpDefaultTokenType,
		Ticker:   "UNMN",
		Name:     "Unmined Token",
		Quantity: 5,
	})
	if err != nil {
		return fmt.Errorf("error building unmined genesis; %w", err)
	}
	if err := saver.NewTxMinimal(false).SaveTxs(s.ctx, dbi.WireBlockToBlock(
		memo.GetBlockFromTxs([]*wire.MsgTx{umGenesisTx.MsgTx}, nil))); err != nil {
		return fmt.Errorf("error saving unmined genesis minimally; %w", err)
	}
	umBackfill := act_maint.NewSlpValidityBackfill(s.ctx, false)
	if err := umBackfill.Run(); err != nil {
		return fmt.Errorf("error running slp validity backfill with unmined candidate; %w", err)
	}
	if umBackfill.MempoolTail != 1 {
		return fmt.Errorf("error backfill counted %d unmined candidates, expected 1", umBackfill.MempoolTail)
	}
	if umBackfill.SlpTxs != 0 || umBackfill.Valid != 0 || umBackfill.Invalid != 0 || umBackfill.Pending != 0 {
		return fmt.Errorf("error backfill touched txs with only an unmined candidate: slp %d valid %d invalid %d pending %d",
			umBackfill.SlpTxs, umBackfill.Valid, umBackfill.Invalid, umBackfill.Pending)
	}
	if err := s.checkStatus("unmined candidate skipped", umGenesisTx, slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
		return err
	}
	var umGenesisRow = &item_slp.Genesis{TxHash: txHash32(umGenesisTx)}
	if err := db.GetItem(s.ctx, umGenesisRow); err == nil {
		return fmt.Errorf("error unmined genesis was transcribed, expected skipped")
	} else if !errors.Is(err, client.EntryNotFoundError) {
		return fmt.Errorf("error checking unmined genesis transcription; %w", err)
	}
	if err := s.mine(700003, 3, umGenesisTx); err != nil {
		return err
	}
	umBackfill2 := act_maint.NewSlpValidityBackfill(s.ctx, false)
	if err := umBackfill2.Run(); err != nil {
		return fmt.Errorf("error running slp validity backfill after mining; %w", err)
	}
	if umBackfill2.MempoolTail != 0 {
		return fmt.Errorf("error backfill still skipped %d as unmined after mining", umBackfill2.MempoolTail)
	}
	if umBackfill2.SlpTxs != 1 || umBackfill2.Valid != 1 {
		return fmt.Errorf("error backfill after mining validated slp %d valid %d, expected 1 and 1",
			umBackfill2.SlpTxs, umBackfill2.Valid)
	}
	if err := s.checkStatus("mined on re-run", umGenesisTx, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	if err := db.GetItem(s.ctx, umGenesisRow); err != nil {
		return fmt.Errorf("error mined genesis row not transcribed by backfill; %w", err)
	}
	log.Printf("✓ backfill skipped the unmined candidate and decided it once mined\n")

	// Zero-quantity send: a legal burn/no-op — declared output sum 0 needs no
	// token inputs, so it is self-evidently valid. The live path must write
	// the send row and a VALID verdict while writing NO token output rows;
	// this is the tx-level shape the per-output graph fields cannot show
	_, zeroFundUtxo, err := s.fund(1e5)
	if err != nil {
		return err
	}
	zeroSend, err := build.TokenSend(build.TokenSendRequest{
		Wallet:    s.wallet(zeroFundUtxo),
		TokenHash: bfTokenHash,
		Recipient: test_tx.Address2,
		Quantity:  0,
		TokenType: memo.SlpDefaultTokenType,
	})
	if err != nil {
		return fmt.Errorf("error building zero-quantity send; %w", err)
	}
	if err := s.save(zeroSend); err != nil {
		return err
	}
	if err := s.checkStatus("zero-quantity send", zeroSend, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	zeroSendHash := txHash32(zeroSend)
	var zeroSendRow = &item_slp.Send{TxHash: zeroSendHash}
	if err := db.GetItem(s.ctx, zeroSendRow); err != nil {
		return fmt.Errorf("error zero-quantity send row not transcribed; %w", err)
	}
	if !bytes.Equal(zeroSendRow.TokenHash[:], bfTokenHash) {
		return fmt.Errorf("error zero-quantity send row token %x, expected %x", zeroSendRow.TokenHash, bfTokenHash)
	}
	zeroOutputs, err := item_slp.GetOutputs(s.ctx, []memo.Out{
		{TxHash: zeroSendHash[:], Index: 1},
		{TxHash: zeroSendHash[:], Index: 2},
	})
	if err != nil {
		return fmt.Errorf("error getting zero-quantity send output rows; %w", err)
	}
	if len(zeroOutputs) != 0 {
		return fmt.Errorf("error zero-quantity send has %d token output rows, expected none", len(zeroOutputs))
	}
	log.Printf("✓ zero-quantity send: valid verdict, send row, no token output rows\n")

	// Unparseable vout-0 lokad: carries the SLP prefix, so the backfill must
	// consider it, but only 3 pushes — below the minimum for any action.
	// Transcription declines it and strict validation records INVALID: a
	// verdict row with no action row
	_, malformedFundUtxo, err := s.fund(1e5)
	if err != nil {
		return err
	}
	// OP_RETURN, push "SLP\x00", push 0x01 (type 1), push "SEND"
	malformedTx, err := build.SimpleSingle(s.wallet(malformedFundUtxo), []*memo.Output{
		{Script: rawScript{pk: []byte{0x6a, 0x04, 'S', 'L', 'P', 0x00, 0x01, 0x01, 0x04, 'S', 'E', 'N', 'D'}}},
		gen.GetAddressOutput(test_tx.Address2, memo.DustMinimumOutput),
	})
	if err != nil {
		return fmt.Errorf("error building malformed slp tx; %w", err)
	}
	if err := saver.NewTxMinimal(false).SaveTxs(s.ctx, dbi.WireBlockToBlock(
		memo.GetBlockFromTxs([]*wire.MsgTx{malformedTx.MsgTx}, nil))); err != nil {
		return fmt.Errorf("error saving malformed slp tx minimally; %w", err)
	}
	if err := s.mine(700004, 4, malformedTx); err != nil {
		return err
	}
	if err := act_maint.NewSlpValidityBackfill(s.ctx, false).Run(); err != nil {
		return fmt.Errorf("error running backfill for malformed slp tx; %w", err)
	}
	if err := s.checkStatus("malformed vout-0 lokad", malformedTx,
		slp_tx.StatusInvalid, slp_tx.ReasonParse); err != nil {
		return err
	}
	malformedHash := txHash32(malformedTx)
	if err := db.GetItem(s.ctx, &item_slp.Send{TxHash: malformedHash}); err == nil {
		return fmt.Errorf("error malformed slp tx has a send row, expected none")
	} else if !errors.Is(err, client.EntryNotFoundError) {
		return fmt.Errorf("error checking malformed slp send row; %w", err)
	}
	log.Printf("✓ malformed lokad: invalid verdict with no action row\n")

	// Tx-level slp attach: validity is a per-tx fact, so the graph layer must
	// surface it from the action + validity rows alone — including the two
	// sparse shapes above, which have no token output rows to hang it on
	nonSlpTx, _, err := s.fund(1e5)
	if err != nil {
		return err
	}
	var attachTxs = []*model.Tx{
		{Hash: model.Hash(txHash32(bfGenesisTx))},
		{Hash: model.Hash(txHash32(fakeSend))},
		{Hash: model.Hash(txHash32(nonSlpTx))},
		{Hash: model.Hash(zeroSendHash)},
		{Hash: model.Hash(malformedHash)},
	}
	if err := attach.ToTxs(s.ctx, []attach.Field{{Name: "slp", Fields: []attach.Field{
		{Name: "type"}, {Name: "token_hash"}, {Name: "validity"},
		{Name: "genesis", Fields: []attach.Field{{Name: "ticker"}}},
	}}}, attachTxs); err != nil {
		return fmt.Errorf("error attaching slp to txs; %w", err)
	}
	genesisSlp, fakeSendSlp, nonSlpSlp := attachTxs[0].Slp, attachTxs[1].Slp, attachTxs[2].Slp
	if genesisSlp == nil || genesisSlp.Type == nil || *genesisSlp.Type != model.SlpActionTypeGenesis ||
		genesisSlp.TokenHash == nil || *genesisSlp.TokenHash != model.Hash(txHash32(bfGenesisTx)) ||
		genesisSlp.Validity != model.SlpValidityValid {
		return fmt.Errorf("error tx-level slp for genesis = %+v, expected valid genesis of itself", genesisSlp)
	}
	if fakeSendSlp == nil || fakeSendSlp.Type == nil || *fakeSendSlp.Type != model.SlpActionTypeSend ||
		fakeSendSlp.TokenHash == nil || !bytes.Equal(fakeSendSlp.TokenHash[:], tokenHash) ||
		fakeSendSlp.Validity != model.SlpValidityInvalid {
		return fmt.Errorf("error tx-level slp for fake send = %+v, expected invalid send of token A", fakeSendSlp)
	}
	if nonSlpSlp != nil {
		return fmt.Errorf("error tx-level slp for plain bch tx = %+v, expected nil", nonSlpSlp)
	}
	zeroSlp, malformedSlp := attachTxs[3].Slp, attachTxs[4].Slp
	if zeroSlp == nil || zeroSlp.Type == nil || *zeroSlp.Type != model.SlpActionTypeSend ||
		zeroSlp.TokenHash == nil || !bytes.Equal(zeroSlp.TokenHash[:], bfTokenHash) ||
		zeroSlp.Validity != model.SlpValidityValid {
		return fmt.Errorf("error tx-level slp for zero-quantity send = %+v, expected valid send", zeroSlp)
	}
	if malformedSlp == nil || malformedSlp.Type != nil || malformedSlp.TokenHash != nil ||
		malformedSlp.Validity != model.SlpValidityInvalid {
		return fmt.Errorf("error tx-level slp for malformed lokad = %+v, expected invalid with nil type", malformedSlp)
	}
	// Genesis resolution: a genesis action resolves to itself, a send to its
	// token's genesis; no token hash (malformed) or plain BCH means none
	if genesisSlp.Genesis == nil || !bytes.Equal(genesisSlp.Genesis.Hash[:], bfTokenHash) {
		return fmt.Errorf("error tx-level slp genesis for genesis action = %+v, expected itself", genesisSlp.Genesis)
	}
	if zeroSlp.Genesis == nil || zeroSlp.Genesis.Ticker != "BKFL" {
		return fmt.Errorf("error tx-level slp genesis for zero-quantity send = %+v, expected BKFL token", zeroSlp.Genesis)
	}
	if malformedSlp.Genesis != nil {
		return fmt.Errorf("error tx-level slp genesis for malformed lokad = %+v, expected nil", malformedSlp.Genesis)
	}
	log.Printf("✓ tx-level slp attach surfaces action and verdict without output rows\n")
	return nil
}
