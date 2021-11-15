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
	"github.com/memocash/server/ref/bitcoin/util/testing/test_block"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/server/ref/dbi"
)

const (
	FundingValue = 100000
	SendAmount   = 99000
	SendAmount2  = 98000
	SendAmount3  = 97000
)

type DoubleSpend struct {
	TxSaver         dbi.TxSave
	BlockSaver      dbi.BlockSave
	FundingPkScript []byte
}

func (s *DoubleSpend) Init(wallet *build.Wallet) error {
	s.TxSaver = saver.CombinedTxSaver(false)
	s.BlockSaver = saver.NewBlock(false)
	fundingTx, err := test_tx.GetFundingTx(wallet.Address, FundingValue)
	if err != nil {
		return jerr.Get("error getting funding tx for address", err)
	}
	if err := s.SaveBlock([]*memo.Tx{fundingTx}); err != nil {
		return jerr.Get("error saving funding tx", err)
	}
	wallet.Getter.AddChangeUTXO(script.GetOutputUTXOs(fundingTx)[0])
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

func (s *DoubleSpend) SaveBlock(txs []*memo.Tx) error {
	var wireTxs = make([]*wire.MsgTx, len(txs))
	for i := range txs {
		wireTxs[i] = txs[i].MsgTx
	}
	block := test_block.GetNextBlock(wireTxs)
	if err := s.TxSaver.SaveTxs(block); err != nil {
		return jerr.Get("error adding txs block to network", err)
	}
	if err := s.BlockSaver.SaveBlock(block.Header); err != nil {
		return jerr.Get("error saving block header for double spend grp", err)
	}
	return nil
}

func (s *DoubleSpend) GetAddressBalance(address string) (int64, error) {
	balance, err := get.NewBalanceFromAddress(address)
	if err != nil {
		return 0, jerr.Get("error getting address from string for double spend balance", err)
	}
	if err := balance.GetBalance(); err != nil {
		return 0, jerr.Get("error getting address balance by utxos from network", err)
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
