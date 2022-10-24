package grp

import (
	"bytes"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/obj/get"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_block"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
)

const (
	FundingValue = 100000
	SendAmount   = 99000
	SendAmount2  = 1000
)

type DoubleSpend struct {
	TxSaver         dbi.TxSave
	DelayedTxSaver  dbi.TxSave
	DelayAmount     int
	BlockSaver      dbi.BlockSave
	FundingPkScript []byte
	OldBlocks       []*wire.MsgBlock
}

func (s *DoubleSpend) Init(wallet *build.Wallet) error {
	s.TxSaver = saver.NewCombinedAll(false)
	s.DelayedTxSaver = saver.NewClearSuspect()
	s.DelayAmount = int(config.GetBlocksToConfirm())
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

type CreateTx struct {
	Address  wallet.Address
	Quantity int64
	Wallet   build.Wallet
	Receive  *build.Wallet
	MemoTx   *memo.Tx
}

func (s *DoubleSpend) CreateTxs(txs []*CreateTx) error {
	var err error
	for i, tx := range txs {
		if tx.MemoTx, err = s.Create(gen.GetAddressOutput(tx.Address, tx.Quantity), tx.Wallet); err != nil {
			return jerr.Getf(err, "error saving to address: %s %d", tx.Address, tx.Quantity)
		}
		if tx.Receive != nil {
			tx.Receive.Getter.AddChangeUTXO(script.GetOutputUTXOs(tx.MemoTx)[0])
		}
		jlog.Logf("tx %d: %s %d %s\n", i, tx.Address.GetEncoded(), tx.Quantity, hs.GetTxString(tx.MemoTx.GetHash()))
	}
	return nil
}

func (s *DoubleSpend) SaveBlock(txs []*memo.Tx) error {
	var wireTxs = make([]*wire.MsgTx, len(txs))
	for i := range txs {
		wireTxs[i] = txs[i].MsgTx
	}
	block := test_block.GetNextBlock(wireTxs)
	if err := s.BlockSaver.SaveBlock(block.Header); err != nil {
		return jerr.Get("error saving block header for double spend grp", err)
	}
	if err := s.TxSaver.SaveTxs(block); err != nil {
		return jerr.Get("error adding txs block to network", err)
	}
	if len(s.OldBlocks) > s.DelayAmount {
		if err := s.DelayedTxSaver.SaveTxs(s.OldBlocks[len(s.OldBlocks)-s.DelayAmount-1]); err != nil {
			return jerr.Get("error adding delayed txs block to network", err)
		}
	}
	s.OldBlocks = append(s.OldBlocks, block)
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

type AddressBalance struct {
	Address  string
	Expected int64
}

func (s *DoubleSpend) CheckAddressBalances(addressBalances []AddressBalance) error {
	for _, addressBalance := range addressBalances {
		if err := s.CheckAddressBalance(addressBalance.Address, addressBalance.Expected); err != nil {
			return jerr.Getf(err, "error address balance does not match expected: %s %d",
				addressBalance.Address, addressBalance.Expected)
		}
	}
	return nil
}

type TxSuspect struct {
	Tx       []byte
	Expected bool
}

func (s *DoubleSpend) CheckSuspects(checkSuspects []TxSuspect) error {
	var txHashes = make([][]byte, len(checkSuspects))
	for i := range checkSuspects {
		txHashes[i] = checkSuspects[i].Tx
	}
	txSuspects, err := item.GetTxSuspects(txHashes)
	if err != nil {
		return jerr.Get("error getting tx suspects for double spend check", err)
	}
	for i, checkSuspect := range checkSuspects {
		var txSuspectFound bool
		for _, txSuspect := range txSuspects {
			if bytes.Equal(txSuspect.TxHash, checkSuspect.Tx) {
				txSuspectFound = true
				break
			}
		}
		if checkSuspect.Expected != txSuspectFound {
			return jerr.Newf("error check suspect expected does not match actual: %d %s %t %t",
				i, hs.GetTxString(checkSuspect.Tx), checkSuspect.Expected, txSuspectFound)
		}
	}
	return nil
}
