package tasks

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/server/ref/config"
	"github.com/memocash/server/test/grp"
	"github.com/memocash/server/test/suite"
)

var doubleSpendTest = suite.Test{
	Name: TestDoubleSpend,
	Test: func(r *suite.TestRequest) error {
		var doubleSpend = &grp.DoubleSpend{}
		address1Wallet := test_tx.GetKeyWallet(&test_tx.Address1key, nil)
		if err := doubleSpend.Init(&address1Wallet); err != nil {
			return jerr.Get("error initializing double spend group", err)
		}
		address1WalletCopy := test_tx.CopyTestWallet(address1Wallet)
		address2Wallet := test_tx.GetKeyWallet(&test_tx.Address2key, nil)
		address3Wallet := test_tx.GetKeyWallet(&test_tx.Address3key, nil)
		var tx0 = &grp.CreateTx{Address: test_tx.Address2, Quantity: grp.SendAmount, Wallet: address1Wallet,
			Receive: &address2Wallet}
		var tx1 = &grp.CreateTx{Address: test_tx.Address3, Quantity: grp.SendAmount, Wallet: address1WalletCopy,
			Receive: &address3Wallet}
		var tx2 = &grp.CreateTx{Address: test_tx.Address4, Quantity: grp.SendAmount2, Wallet: address2Wallet}
		var tx3 = &grp.CreateTx{Address: test_tx.Address5, Quantity: grp.SendAmount2, Wallet: address3Wallet}
		if err := doubleSpend.CreateTxs([]*grp.CreateTx{tx0, tx1, tx2, tx3}); err != nil {
			return jerr.Get("error creating double spend initial transactions", err)
		}
		var address1Expected = grp.FundingValue - grp.SendAmount - memo.FeeP2pkh1In2OutTx
		var address2expected = grp.SendAmount - grp.SendAmount2 - memo.FeeP2pkh1In2OutTx
		if err := doubleSpend.CheckAddressBalances([]grp.AddressBalance{
			{Address: test_tx.Address1String, Expected: address1Expected},
			{Address: test_tx.Address2String, Expected: address2expected},
			{Address: test_tx.Address3String, Expected: 0},
			{Address: test_tx.Address4String, Expected: grp.SendAmount2},
			{Address: test_tx.Address5String, Expected: 0},
		}); err != nil {
			return jerr.Get("error checking address balances for double spend test before block", err)
		}
		if err := doubleSpend.SaveBlock([]*memo.Tx{tx1.MemoTx, tx3.MemoTx}); err != nil {
			return jerr.Get("error saving address 3 tx block", err)
		}
		if err := doubleSpend.CheckAddressBalances([]grp.AddressBalance{
			{Address: test_tx.Address1String, Expected: address1Expected},
			{Address: test_tx.Address2String, Expected: 0},
			{Address: test_tx.Address3String, Expected: address2expected},
			{Address: test_tx.Address4String, Expected: 0},
			{Address: test_tx.Address5String, Expected: grp.SendAmount2},
		}); err != nil {
			return jerr.Get("error checking address balances for double spend test after block", err)
		}
		if err := doubleSpend.CheckSuspects([]grp.TxSuspect{
			{Tx: tx0.MemoTx.GetHash(), Expected: true},
			{Tx: tx1.MemoTx.GetHash(), Expected: true},
			{Tx: tx2.MemoTx.GetHash(), Expected: true},
			{Tx: tx3.MemoTx.GetHash(), Expected: true},
		}); err != nil {
			return jerr.Get("error checking tx suspects for double spend test", err)
		}
		defaultBlocksToConfirm := int(config.GetBlocksToConfirm())
		for i := 0; i <= defaultBlocksToConfirm; i++ {
			var txA = &grp.CreateTx{Address: test_tx.Address4, Quantity: grp.SendAmount2, Wallet: address2Wallet}
			var txB = &grp.CreateTx{Address: test_tx.Address5, Quantity: grp.SendAmount2, Wallet: address3Wallet}
			if err := doubleSpend.CreateTxs([]*grp.CreateTx{txA, txB}); err != nil {
				return jerr.Getf(err, "error creating double spend txA/txB txs: %d", i)
			}
			if err := doubleSpend.SaveBlock([]*memo.Tx{txB.MemoTx}); err != nil {
				return jerr.Getf(err, "error saving txB block %d", i)
			}
		}
		if err := doubleSpend.CheckSuspects([]grp.TxSuspect{
			{Tx: tx0.MemoTx.GetHash(), Expected: true},
			{Tx: tx1.MemoTx.GetHash(), Expected: false},
			{Tx: tx2.MemoTx.GetHash(), Expected: true},
			{Tx: tx3.MemoTx.GetHash(), Expected: false},
		}); err != nil {
			return jerr.Get("error checking tx suspects for double spend test after block", err)
		}
		// TODO: Check txAs ARE marked lost
		// TODO: Check txBs ARE NOT marked suspect
		return nil
	},
}
