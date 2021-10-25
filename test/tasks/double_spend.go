package tasks

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/node/obj/get"
	"github.com/memocash/server/node/obj/saver"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/tx/parse"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/server/test/suite"
)

var doubleSpendTest = suite.Test{
	Name: TestDoubleSpend,
	Test: func(r *suite.TestRequest) error {
		const (
			FundingValue = 1e8
			SendAmount   = 1e5
		)
		fundingTx, err := test_tx.GetFundingTx(test_tx.Address1, FundingValue)
		if err != nil {
			return jerr.Get("error getting funding tx for address", err)
		}
		txSaver := saver.CombinedTxSaver(false)
		if err := txSaver.SaveTxs(memo.GetBlockFromTxs([]*wire.MsgTx{fundingTx.MsgTx}, nil)); err != nil {
			return jerr.Get("error saving funding tx", err)
		}
		jlog.Logf("fundingTx: %s\n", fundingTx.MsgTx.TxHash())
		parse.GetTxInfo(fundingTx).Print()
		address1Wallet := test_tx.GetKeyWallet(&test_tx.Address1key, script.GetOutputUTXOs(fundingTx))
		pkScript, err := fundingTx.Outputs[0].Script.Get()
		if err != nil {
			return jerr.Get("error getting output script", err)
		}
		txRequest := gen.TxRequest{
			Outputs: []*memo.Output{
				gen.GetAddressOutput(test_tx.Address2, SendAmount),
			},
			InputsToUse: []memo.UTXO{{
				Input: memo.TxInput{
					PkScript:     pkScript,
					PkHash:       test_tx.Address1pkHash,
					Value:        FundingValue,
					PrevOutHash:  fundingTx.GetHash(),
					PrevOutIndex: 0,
				},
			}},
			Change:  address1Wallet.GetChange(),
			KeyRing: address1Wallet.KeyRing,
		}
		tx2, err := gen.Tx(txRequest)
		if err != nil {
			return jerr.Get("error generating transaction 2", err)
		}
		if err := txSaver.SaveTxs(memo.GetBlockFromTxs([]*wire.MsgTx{tx2.MsgTx}, nil)); err != nil {
			return jerr.Get("error saving tx2", err)
		}
		jlog.Logf("tx2: %s\n", tx2.MsgTx.TxHash())
		parse.GetTxInfo(tx2).Print()
		txRequest.Outputs = []*memo.Output{
			gen.GetAddressOutput(test_tx.Address3, SendAmount),
		}
		tx3, err := gen.Tx(txRequest)
		if err != nil {
			return jerr.Get("error generating transaction 3", err)
		}
		if err := txSaver.SaveTxs(memo.GetBlockFromTxs([]*wire.MsgTx{tx3.MsgTx}, nil)); err != nil {
			return jerr.Get("error saving tx3", err)
		}
		jlog.Logf("tx3: %s\n", tx3.MsgTx.TxHash())
		parse.GetTxInfo(tx3).Print()
		txBlock := memo.GetBlockFromTxs([]*wire.MsgTx{fundingTx.MsgTx, tx3.MsgTx}, &test_tx.Block1Header)
		if err := txSaver.SaveTxs(txBlock); err != nil {
			return jerr.Get("error adding tx1 tx3 block1 to network", err)
		}
		newBalance, err := get.NewBalanceFromAddress(test_tx.Address1String)
		if err != nil {
			return jerr.Get("error getting address from string for balance", err)
		}
		if err := newBalance.GetUtxos(); err != nil {
			return jerr.Get("error getting address 1 balance from network", err)
		} else if newBalance.Balance > FundingValue {
			return jerr.Newf("error address 1 balance greater than funding amount: %s",
				jfmt.AddCommas(newBalance.Balance))
		}
		//jlog.Logf("newBalance.Balance address1: %s\n", jfmt.AddCommas(newBalance.Balance))
		newBalance2, err := get.NewBalanceFromAddress(test_tx.Address2String)
		if err != nil {
			return jerr.Get("error getting address 2 from string for balance", err)
		}
		if err := newBalance2.GetUtxos(); err != nil {
			return jerr.Get("error getting address 2 balance from network", err)
		} else if newBalance2.Balance != SendAmount {
			return jerr.Newf("error address 2 balance not equal send amount: %s",
				jfmt.AddCommas(newBalance2.Balance))
		}
		//jlog.Logf("newBalance.Balance address2: %s\n", jfmt.AddCommas(newBalance.Balance))
		newBalance3, err := get.NewBalanceFromAddress(test_tx.Address3String)
		if err != nil {
			return jerr.Get("error getting address 3 from string for balance", err)
		}
		if err := newBalance3.GetUtxos(); err != nil {
			return jerr.Get("error getting address 3 balance from network", err)
		} else if newBalance3.Balance != 0 {
			return jerr.Newf("error address 3 balance not equal 0: %s",
				jfmt.AddCommas(newBalance3.Balance))
		}
		//jlog.Logf("newBalance.Balance address3: %s\n", jfmt.AddCommas(newBalance.Balance))
		return nil
	},
}
