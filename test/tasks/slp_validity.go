package tasks

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	item_slp "github.com/memocash/index/db/item/slp"
	act_maint "github.com/memocash/index/node/act/maint"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
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
	// resolve as processed-not-SLP rather than unknown
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
	return nil
}
