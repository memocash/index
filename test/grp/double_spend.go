package grp

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/memocash/server/node/obj/get"
	"github.com/memocash/server/node/obj/saver"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/build"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/server/ref/dbi"
)

const (
	FundingValue = 1e8
	SendAmount   = 1e5
	SendAmount2  = 1e4
)

type DoubleSpend struct {
	TxSaver         dbi.TxSave
	BlockSaver      dbi.BlockSave
	FundingTx       *memo.Tx
	FundingPkScript []byte
}

func (s *DoubleSpend) Init(wallet *build.Wallet) error {
	s.TxSaver = saver.CombinedTxSaver(false)
	s.BlockSaver = saver.NewBlock(false)
	var err error
	if s.FundingTx, err = test_tx.GetFundingTx(wallet.Address, FundingValue); err != nil {
		return jerr.Get("error getting funding tx for address", err)
	}
	if err := s.TxSaver.SaveTxs(memo.GetBlockFromTxs([]*wire.MsgTx{s.FundingTx.MsgTx}, nil)); err != nil {
		return jerr.Get("error saving funding tx", err)
	}
	wallet.Getter.AddChangeUTXO(script.GetOutputUTXOs(s.FundingTx)[0])
	return nil
}

func (s *DoubleSpend) Create(output *memo.Output, wallet build.Wallet) (*memo.Tx, error) {
	var txRequest = gen.TxRequest{
		Outputs: []*memo.Output{output},
		Getter:  wallet.Getter,
		Change:  wallet.GetChange(),
		KeyRing: wallet.KeyRing,
	}
	tx, err := gen.Tx(txRequest)
	if err != nil {
		return nil, jerr.Get("error generating transaction", err)
	}
	if err := s.TxSaver.SaveTxs(memo.GetBlockFromTxs([]*wire.MsgTx{tx.MsgTx}, nil)); err != nil {
		return nil, jerr.Get("error saving tx", err)
	}
	return tx, nil
}

func (s *DoubleSpend) SaveBlock(tx *memo.Tx) error {
	txBlock := memo.GetBlockFromTxs([]*wire.MsgTx{s.FundingTx.MsgTx, tx.MsgTx}, &test_tx.Block1Header)
	if err := s.TxSaver.SaveTxs(txBlock); err != nil {
		return jerr.Get("error adding tx1 tx3 block1 to network", err)
	}
	if err := s.BlockSaver.SaveBlock(test_tx.Block1Header); err != nil {
		return jerr.Get("error saving block header 1 for double spend grp", err)
	}
	return nil
}

func (s *DoubleSpend) GetAddressBalance(address string) (int64, error) {
	balance, err := get.NewBalanceFromAddress(address)
	if err != nil {
		return 0, jerr.Get("error getting address 2 from string for balance", err)
	}
	if err := balance.GetBalance(); err != nil {
		return 0, jerr.Get("error getting address 2 balance from network", err)
	}
	return balance.Balance, nil
}

func (s *DoubleSpend) CheckAddressBalance(address string, expectedBalance int64) error {
	balance, err := s.GetAddressBalance(address)
	if err != nil {
		return jerr.Get("error getting balance for address", err)
	}
	if balance != expectedBalance {
		return jerr.Newf("error double spend balance does not equal expected: %s %s",
			jfmt.AddCommas(balance), jfmt.AddCommas(expectedBalance))
	}
	return nil
}
