package tasks

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	item_slp "github.com/memocash/index/db/item/slp"
	act_maint "github.com/memocash/index/node/act/maint"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	slp_tx "github.com/memocash/index/ref/bitcoin/tx/slp"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"github.com/memocash/index/test/suite"
	"log"
)

// SlpValidity is an end-to-end test of SLP validity verdicts against real
// queue shards: a token lifecycle (genesis, send, mint) is fed through the
// full saver pipeline, alongside a fake send and fake mint with no valid
// token inputs, out-of-order (child before parent) arrival, the
// slp-validity-sweep index and audit backfills, and NFT1 group/child geneses.
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
	if err := act_maint.NewSlpValiditySweep(s.ctx, false, false).Run(); err != nil {
		return fmt.Errorf("error running slp validity index sweep; %w", err)
	}
	for _, tx := range backfillTxs {
		if err := s.checkStatus("backfill re-derived verdict", tx, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
			return err
		}
	}
	log.Printf("✓ index sweep re-derived wiped verdicts\n")

	// Audit backfill: strip a tx's verdict and transcription rows so the
	// index sweep cannot enumerate it (simulating a tx the live path never
	// transcribed), then verify only the tx-output audit finds it
	var gapTx = chainTxs[3] // last generation: nothing decided cascades to it
	var gapHash = txHash32(gapTx)
	if err := db.Remove([]db.Object{
		&item_slp.Validity{TxHash: gapHash},
		&item_slp.Send{TxHash: gapHash},
		&item_slp.Output{TxHash: gapHash, Index: 1},
		&item_slp.Output{TxHash: gapHash, Index: 2},
	}); err != nil {
		return fmt.Errorf("error removing gap tx rows for audit simulation; %w", err)
	}
	if err := act_maint.NewSlpValiditySweep(s.ctx, false, false).Run(); err != nil {
		return fmt.Errorf("error running slp validity index sweep with gap tx; %w", err)
	}
	if err := s.checkStatus("gap tx invisible to index sweep", gapTx, slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
		return err
	}
	// Legacy cursor guard: the pre-dataset sweeper stored an 8-byte height
	// under the same process status, which the audit must not resume from.
	// A realistic legacy height (leading zero bytes) sorts before nearly
	// every output uid, so seed the worst case instead: an 8-byte cursor
	// sorting after every uid, which would skip the whole shard if resumed
	for _, shardConfig := range config.GetQueueShards() {
		legacy := item.NewProcessStatus(uint(shardConfig.Shard), item.ProcessStatusSlpValiditySweep)
		legacy.Status = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		if err := legacy.Save(); err != nil {
			return fmt.Errorf("error saving legacy cursor for shard %d; %w", shardConfig.Shard, err)
		}
	}
	if err := act_maint.NewSlpValiditySweep(s.ctx, false, true).Run(); err != nil {
		return fmt.Errorf("error running slp validity audit; %w", err)
	}
	if err := s.checkStatus("gap tx found by audit", gapTx, slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	sendHashes, err := item_slp.GetTopicTxHashes(s.ctx, db.TopicSlpSend, db.GetShardIdFromByte32(gapHash[:]), nil)
	if err != nil {
		return fmt.Errorf("error getting slp send hashes after audit; %w", err)
	}
	var sendRestored bool
	for _, sendHash := range sendHashes {
		if sendHash == gapHash {
			sendRestored = true
			break
		}
	}
	if !sendRestored {
		return fmt.Errorf("error gap tx slp send row not restored by audit")
	}
	// A completed audit clears its per-shard resume cursors
	for _, shardConfig := range config.GetQueueShards() {
		status, err := item.GetProcessStatus(s.ctx, uint(shardConfig.Shard), item.ProcessStatusSlpValiditySweep)
		if err != nil {
			return fmt.Errorf("error getting audit cursor for shard %d; %w", shardConfig.Shard, err)
		}
		if len(status.Status) != 0 {
			return fmt.Errorf("error audit cursor for shard %d not cleared: %x", shardConfig.Shard, status.Status)
		}
	}
	log.Printf("✓ audit transcribed and validated the gap tx; cursors cleared\n")

	// Repopulation window: the wipe-and-repopulate deploy leaves chain rows
	// intact while slp rows are missing until the async repopulation reaches
	// them. Simulate a send whose token's genesis is not yet repopulated: the
	// genesis's transcription rows and the pair's verdicts are stripped, and
	// the index sweep must leave the send pending - never invalid - until the
	// genesis is re-transcribed (audit), whose verdict then cascades to it.
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
	if err := act_maint.NewSlpValiditySweep(s.ctx, false, false).Run(); err != nil {
		return fmt.Errorf("error running slp validity index sweep with missing genesis; %w", err)
	}
	if err := s.checkStatus("send with unrepopulated genesis stays pending", send1,
		slp_tx.StatusPending, slp_tx.ReasonNone); err != nil {
		return err
	}
	if err := act_maint.NewSlpValiditySweep(s.ctx, false, true).Run(); err != nil {
		return fmt.Errorf("error running slp validity audit to repopulate genesis; %w", err)
	}
	if err := s.checkStatus("genesis repopulated by audit", genesisTx,
		slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	if err := s.checkStatus("send resolved after genesis repopulated", send1,
		slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	log.Printf("✓ missing genesis left send pending until repopulated\n")

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

	// Audit re-poison guard: strip the poison tx's derived rows + verdict and
	// run the audit backfill. The tx carries a token-A message at vout 0 and a
	// token-B message at a later output; the audit must re-transcribe only the
	// vout-0 message (the transcribe() vout-0 gate), restoring token-A rows
	// and never re-poisoning token-B rows. This exercises the recovery path,
	// which the live-handler poison assertion above does not.
	poisonHash := txHash32(poisonTx)
	if err := db.Remove([]db.Object{
		&item_slp.Send{TxHash: poisonHash},
		&item_slp.Output{TxHash: poisonHash, Index: 1},
		&item_slp.Output{TxHash: poisonHash, Index: 2},
		&item_slp.Validity{TxHash: poisonHash},
	}); err != nil {
		return fmt.Errorf("error removing poison tx rows for audit guard; %w", err)
	}
	if err := act_maint.NewSlpValiditySweep(s.ctx, false, true).Run(); err != nil {
		return fmt.Errorf("error running audit for poison guard; %w", err)
	}
	if err := s.checkStatus("poison tx re-validated by audit", poisonTx,
		slp_tx.StatusValid, slp_tx.ReasonNone); err != nil {
		return err
	}
	var poisonSendAudit = &item_slp.Send{TxHash: poisonHash}
	if err := db.GetItem(s.ctx, poisonSendAudit); err != nil {
		return fmt.Errorf("error getting poison tx send row after audit; %w", err)
	}
	if poisonSendAudit.TokenHash != tokenAHash {
		return fmt.Errorf("error audit restored poison send row as token %x, expected token A (audit re-transcribed a later message)",
			poisonSendAudit.TokenHash)
	}
	poisonOutsAudit, err := item_slp.GetOutputs(s.ctx, []memo.Out{
		{TxHash: poisonTx.GetHash(), Index: 1},
		{TxHash: poisonTx.GetHash(), Index: 2},
		{TxHash: poisonTx.GetHash(), Index: 3},
	})
	if err != nil {
		return fmt.Errorf("error getting poison tx outputs after audit; %w", err)
	}
	var tokenAOuts int
	for _, out := range poisonOutsAudit {
		if out.TokenHash == tokenB {
			return fmt.Errorf("error audit re-poisoned output %d with token B", out.Index)
		}
		if out.TokenHash == tokenAHash {
			tokenAOuts++
		}
	}
	if tokenAOuts != 2 {
		return fmt.Errorf("error audit restored %d token-A outputs, expected 2", tokenAOuts)
	}
	log.Printf("✓ audit re-transcribed only the vout-0 message\n")

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
	return nil
}
