package tasks

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/db/item/addr"
	item_cashtoken "github.com/memocash/index/db/item/cashtoken"
	"github.com/memocash/index/graph/attach"
	"github.com/memocash/index/graph/model"
	act_maint "github.com/memocash/index/node/act/maint"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/cashtoken"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/dbi"
	"github.com/memocash/index/test/suite"
)

// CashToken is an end-to-end test of CashTokens indexing against real queue
// shards: a tx with a fungible token output, an NFT output and a plain BCH
// output is fed through the full saver pipeline, then the token rows, the
// address index and the GraphQL attach layer are checked. A second token tx
// saved with only the minimal saver (chain rows, no token or address rows)
// then exercises the populate-cash-token backfill, including its progress
// logging racing the workers and its behaviour when a shard fails.
var CashToken = suite.Test{
	Name: TestCashToken,
	Test: func(r *suite.TestRequest) error {
		s := cashTokenState{
			ctx:     context.Background(),
			txSaver: saver.NewCombinedTx(false),
		}
		return s.run()
	},
}

type cashTokenState struct {
	ctx     context.Context
	txSaver *saver.CombinedTx
}

func (s *cashTokenState) save(txs ...*wire.MsgTx) error {
	block := dbi.WireBlockToBlock(memo.GetBlockFromTxs(txs, nil))
	if err := s.txSaver.SaveTxs(s.ctx, block); err != nil {
		return fmt.Errorf("error saving txs; %w", err)
	}
	return nil
}

func p2pkhScript(pkHash []byte) []byte {
	return append(append([]byte{txscript.OP_DUP, txscript.OP_HASH160, txscript.OP_DATA_20}, pkHash...),
		txscript.OP_EQUALVERIFY, txscript.OP_CHECKSIG)
}

func (s *cashTokenState) run() error {
	fundingTx, err := test_tx.GetFundingTx(test_tx.Address1, 1e6)
	if err != nil {
		return fmt.Errorf("error getting funding tx; %w", err)
	}
	if err := s.save(fundingTx.MsgTx); err != nil {
		return err
	}
	// The saver does not verify signatures, so a bare spend of the funding
	// output is enough; what matters is the output scripts.
	var category [32]byte
	copy(category[:], fundingTx.GetHash())
	fungible := cashtoken.Token{Category: category, Amount: 5000}
	nft := cashtoken.Token{Category: category, HasNft: true, Capability: cashtoken.CapabilityMinting,
		Commitment: []byte{0xde, 0xad, 0xbe, 0xef}, Amount: 7}
	fundingHash := fundingTx.MsgTx.TxHash()
	tokenTx := wire.NewMsgTx(wire.TxVersion)
	tokenTx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(&fundingHash, 0), nil))
	tokenTx.AddTxOut(wire.NewTxOut(1000, append(cashtoken.Encode(fungible), p2pkhScript(test_tx.Address2pkHash)...)))
	tokenTx.AddTxOut(wire.NewTxOut(1000, append(cashtoken.Encode(nft), p2pkhScript(test_tx.Address3pkHash)...)))
	tokenTx.AddTxOut(wire.NewTxOut(900000, p2pkhScript(test_tx.Address1pkHash)))
	if err := s.save(tokenTx); err != nil {
		return err
	}
	txHash := [32]byte(tokenTx.TxHash())

	// Token rows: exactly the two token outputs, decoded faithfully.
	tokenOutputs, err := item_cashtoken.GetOutputsByTxHashes(s.ctx, [][32]byte{txHash})
	if err != nil {
		return fmt.Errorf("error getting cash token outputs; %w", err)
	}
	if len(tokenOutputs) != 2 {
		return fmt.Errorf("error expected 2 cash token outputs, got %d", len(tokenOutputs))
	}
	for _, out := range tokenOutputs {
		var want cashtoken.Token
		switch out.Index {
		case 0:
			want = fungible
		case 1:
			want = nft
		default:
			return fmt.Errorf("error unexpected cash token output index %d", out.Index)
		}
		if !bytes.Equal(cashtoken.Encode(out.Token()), cashtoken.Encode(want)) {
			return fmt.Errorf("error output %d token mismatch: got %+v, want %+v", out.Index, out.Token(), want)
		}
		log.Printf("✓ cash token output %d: category %s amount %d nft %t capability %s commitment %x\n",
			out.Index, chainhash.Hash(out.Category), out.Amount, out.HasNft, out.Capability, out.Commitment)
	}

	// Address index: the token outputs' addresses saw the tx despite the prefix.
	for _, address := range []wallet.Address{test_tx.Address2, test_tx.Address3} {
		seenTxs, err := addr.GetSeenTxs(s.ctx, *wallet.GetAddrFromPkHash(address.GetPkHash()), nil, 0)
		if err != nil {
			return fmt.Errorf("error getting seen txs for %s; %w", address.GetEncoded(), err)
		}
		var found bool
		for _, seenTx := range seenTxs {
			found = found || seenTx.TxHash == txHash
		}
		if !found {
			return fmt.Errorf("error token tx not in addr_seen_tx for %s", address.GetEncoded())
		}
		log.Printf("✓ addr_seen_tx has token tx for %s\n", address.GetEncoded())
	}

	// GraphQL: cash_token resolves on token outputs, is nil on the plain one,
	// and lock resolves to the wrapped address.
	outputs := []*model.TxOutput{
		{Hash: txHash, Index: 0}, {Hash: txHash, Index: 1}, {Hash: txHash, Index: 2},
	}
	fields := []attach.Field{{Name: "script"}, {Name: "lock", Fields: []attach.Field{{Name: "address"}}},
		{Name: "cash_token", Fields: []attach.Field{{Name: "category"}, {Name: "amount"}, {Name: "nft"},
			{Name: "capability"}, {Name: "commitment"}}}}
	if err := attach.ToOutputs(s.ctx, fields, outputs); err != nil {
		return fmt.Errorf("error attaching outputs; %w", err)
	}
	if outputs[2].CashToken != nil {
		return fmt.Errorf("error plain output has cash_token: %+v", *outputs[2].CashToken)
	}
	ct := outputs[0].CashToken
	if ct == nil || ct.Category != model.Hash(category) || ct.Amount != 5000 || ct.Nft ||
		ct.Capability != model.CashTokenCapabilityNone || len(ct.Commitment) != 0 {
		return fmt.Errorf("error fungible output cash_token mismatch: %+v", ct)
	}
	ct = outputs[1].CashToken
	if ct == nil || ct.Category != model.Hash(category) || ct.Amount != 7 || !ct.Nft ||
		ct.Capability != model.CashTokenCapabilityMinting || !bytes.Equal(ct.Commitment, nft.Commitment) {
		return fmt.Errorf("error nft output cash_token mismatch: %+v", ct)
	}
	if outputs[0].Lock == nil || outputs[0].Lock.Address != model.Address(*wallet.GetAddrFromPkHash(test_tx.Address2pkHash)) {
		return fmt.Errorf("error fungible output lock mismatch: %+v", outputs[0].Lock)
	}
	log.Printf("✓ graphql cash_token and lock attach\n")
	return s.runBackfill(fundingTx.MsgTx, category)
}

// runBackfill saves a token tx through the minimal saver only, so it has
// chain rows but no cash_token_output or addr_seen_tx rows, then runs the
// backfill and checks it wrote both. The log interval is tiny so the
// progress logger reads checkpoints while workers advance them (run under
// -race to verify). Finally a cancelled context must make Populate return an
// error promptly with all workers joined, rather than hang or leak.
func (s *cashTokenState) runBackfill(fundingTx *wire.MsgTx, category [32]byte) error {
	fundingHash := fundingTx.TxHash()
	backfillTx := wire.NewMsgTx(wire.TxVersion)
	backfillTx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(&fundingHash, 1), nil))
	backfillToken := cashtoken.Token{Category: category, HasNft: true, Capability: cashtoken.CapabilityMutable,
		Commitment: []byte{0x01}}
	backfillTx.AddTxOut(wire.NewTxOut(1000, append(cashtoken.Encode(backfillToken), p2pkhScript(test_tx.Address4pkHash)...)))
	block := dbi.WireBlockToBlock(memo.GetBlockFromTxs([]*wire.MsgTx{backfillTx}, nil))
	if err := saver.NewTxMinimal(false).SaveTxs(s.ctx, block); err != nil {
		return fmt.Errorf("error saving backfill tx with minimal saver; %w", err)
	}
	txHash := [32]byte(backfillTx.TxHash())
	before, err := item_cashtoken.GetOutputsByTxHashes(s.ctx, [][32]byte{txHash})
	if err != nil {
		return fmt.Errorf("error getting cash token outputs before backfill; %w", err)
	}
	if len(before) != 0 {
		return fmt.Errorf("error expected no cash token outputs before backfill, got %d", len(before))
	}
	populate := act_maint.NewPopulateCashToken(s.ctx)
	populate.LogInterval = time.Millisecond
	if err := populate.Populate(true); err != nil {
		return fmt.Errorf("error running populate cash token; %w", err)
	}
	if populate.Tokens != 3 {
		return fmt.Errorf("error expected backfill to see 3 token outputs, got %d", populate.Tokens)
	}
	after, err := item_cashtoken.GetOutputsByTxHashes(s.ctx, [][32]byte{txHash})
	if err != nil {
		return fmt.Errorf("error getting cash token outputs after backfill; %w", err)
	}
	if len(after) != 1 || !bytes.Equal(cashtoken.Encode(after[0].Token()), cashtoken.Encode(backfillToken)) {
		return fmt.Errorf("error backfill cash token output mismatch: %+v", after)
	}
	seenTxs, err := addr.GetSeenTxs(s.ctx, *wallet.GetAddrFromPkHash(test_tx.Address4pkHash), nil, 0)
	if err != nil {
		return fmt.Errorf("error getting seen txs after backfill; %w", err)
	}
	var found bool
	for _, seenTx := range seenTxs {
		found = found || seenTx.TxHash == txHash
	}
	if !found {
		return fmt.Errorf("error backfill did not write addr_seen_tx for token output")
	}
	log.Printf("✓ populate-cash-token backfilled token and address rows (checked %d, tokens %d, saved %d)\n",
		populate.Checked, populate.Tokens, populate.Saved)

	cancelled, cancel := context.WithCancel(s.ctx)
	cancel()
	failing := act_maint.NewPopulateCashToken(cancelled)
	failing.LogInterval = time.Millisecond
	var result = make(chan error, 1)
	go func() { result <- failing.Populate(true) }()
	select {
	case err := <-result:
		if err == nil {
			return fmt.Errorf("error expected populate with cancelled context to fail")
		}
		log.Printf("populate with cancelled context failed as expected: %v\n", err)
		if failing.Saved != 0 {
			return fmt.Errorf("error populate with cancelled context saved %d rows", failing.Saved)
		}
	case <-time.After(10 * time.Second):
		return fmt.Errorf("error populate with cancelled context did not return")
	}
	log.Printf("✓ populate-cash-token returns promptly on shard failure\n")
	return nil
}
