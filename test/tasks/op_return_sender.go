package tasks

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	dbMemo "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/dbi"
	"github.com/memocash/index/test/suite"
)

// OpReturnSender is an end-to-end test of memo sender attribution through the
// full saver pipeline: the sender of a memo action is the FIRST tx input
// whose unlock script yields an address (the memo.cash site rule), and a
// memo tx with no parseable input is skipped with a logged process error
// rather than saved with a bogus address.
var OpReturnSender = suite.Test{
	Name: TestOpReturnSender,
	Test: func(r *suite.TestRequest) error {
		s := opReturnSenderState{
			ctx:     context.Background(),
			txSaver: saver.NewCombinedTx(false),
		}
		return s.run()
	},
}

type opReturnSenderState struct {
	ctx     context.Context
	txSaver *saver.CombinedTx
}

func (s *opReturnSenderState) save(msgTx *wire.MsgTx) error {
	block := dbi.WireBlockToBlock(memo.GetBlockFromTxs([]*wire.MsgTx{msgTx}, nil))
	if err := s.txSaver.SaveTxs(s.ctx, block); err != nil {
		return fmt.Errorf("error saving tx; %w", err)
	}
	return nil
}

// senderUnlock builds a P2PKH-shaped unlock script (71-byte signature push +
// 33-byte pubkey push); the pubkey byte determines the derived address. The
// saver pipeline doesn't verify signatures, so the sig bytes are arbitrary.
func senderUnlock(pubKeyByte byte) []byte {
	unlock := []byte{71}
	unlock = append(unlock, bytes.Repeat([]byte{0x30}, 71)...)
	unlock = append(unlock, 33)
	unlock = append(unlock, bytes.Repeat([]byte{pubKeyByte}, 33)...)
	return unlock
}

// memoPostTx hand-crafts a memo post tx with one input per unlock script
// (fake prevouts; the pipeline doesn't resolve or verify them).
func memoPostTx(message string, unlockScripts ...[]byte) (*wire.MsgTx, error) {
	pkScript, err := script.Post{Message: message}.Get()
	if err != nil {
		return nil, fmt.Errorf("error building post script; %w", err)
	}
	var msgTx = wire.NewMsgTx(wire.TxVersion)
	for i, unlockScript := range unlockScripts {
		var prevHash chainhash.Hash
		prevHash[0] = byte(i + 1)
		msgTx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(&prevHash, uint32(i)), unlockScript))
	}
	msgTx.AddTxOut(wire.NewTxOut(0, pkScript))
	return msgTx, nil
}

func (s *opReturnSenderState) checkPostAddr(name, message string, want wallet.Addr, unlockScripts ...[]byte) error {
	msgTx, err := memoPostTx(message, unlockScripts...)
	if err != nil {
		return err
	}
	if err := s.save(msgTx); err != nil {
		return err
	}
	post, err := dbMemo.GetPost(s.ctx, [32]byte(msgTx.TxHash()))
	if err != nil {
		return fmt.Errorf("error getting post for %s; %w", name, err)
	}
	if post == nil {
		return fmt.Errorf("error %s: post not saved", name)
	}
	if wallet.Addr(post.Addr) != want {
		return fmt.Errorf("error %s: post addr %s, want %s", name, wallet.Addr(post.Addr), want)
	}
	log.Printf("✓ %s: post attributed to %s\n", name, want)
	return nil
}

func (s *opReturnSenderState) run() error {
	unlockA := senderUnlock(1)
	unlockB := senderUnlock(2)
	addrA, err := wallet.GetAddrFromUnlockScript(unlockA)
	if err != nil {
		return fmt.Errorf("error parsing unlock a; %w", err)
	}
	addrB, err := wallet.GetAddrFromUnlockScript(unlockB)
	if err != nil {
		return fmt.Errorf("error parsing unlock b; %w", err)
	}
	if *addrA == *addrB {
		return fmt.Errorf("error fixture unlock scripts must derive distinct addresses")
	}

	// Two parseable inputs: attribution follows the first, in both orders,
	// so a regression back to the old last-input selection fails both ways.
	if err := s.checkPostAddr("post a-then-b", "first input wins", *addrA, unlockA, unlockB); err != nil {
		return err
	}
	if err := s.checkPostAddr("post b-then-a", "first input wins swapped", *addrB, unlockB, unlockA); err != nil {
		return err
	}

	// No parseable input: the memo handler must skip the tx (no post row)
	// and log a process error.
	noAddrTx, err := memoPostTx("no parseable input")
	if err != nil {
		return err
	}
	noAddrTx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(&chainhash.Hash{0xff}, 0), nil))
	if err := s.save(noAddrTx); err != nil {
		return err
	}
	post, err := dbMemo.GetPost(s.ctx, [32]byte(noAddrTx.TxHash()))
	if err != nil {
		return fmt.Errorf("error getting no-addr post; %w", err)
	}
	if post != nil {
		return fmt.Errorf("error no-addr post saved with addr %x, want skipped", post.Addr)
	}
	var processError = &item.ProcessError{TxHash: noAddrTx.TxHash()}
	if err := db.GetItem(s.ctx, processError); err != nil {
		return fmt.Errorf("error getting no-addr process error row; %w", err)
	}
	if !strings.Contains(processError.Error, "could not find input pk hash") {
		return fmt.Errorf("error no-addr process error %q, want input pk hash error", processError.Error)
	}
	log.Printf("✓ post with no parseable input skipped with process error\n")
	return nil
}
