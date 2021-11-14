package tasks

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/tx/hs"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
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
		tx2, err := doubleSpend.Create(gen.GetAddressOutput(test_tx.Address2, grp.SendAmount), address1Wallet)
		if err != nil {
			return jerr.Get("error saving tx2 to address 2", err)
		}
		jlog.Logf("tx2: %s\n", hs.GetTxString(tx2.GetHash()))
		address2Wallet.Getter.AddChangeUTXO(script.GetOutputUTXOs(tx2)[0])
		address3Wallet := test_tx.GetKeyWallet(&test_tx.Address3key, nil)
		tx3, err := doubleSpend.Create(gen.GetAddressOutput(test_tx.Address3, grp.SendAmount), address1WalletCopy)
		if err != nil {
			return jerr.Get("error saving tx3 to address 3", err)
		}
		jlog.Logf("tx3: %s\n", hs.GetTxString(tx3.GetHash()))
		address3Wallet.Getter.AddChangeUTXO(script.GetOutputUTXOs(tx3)[0])
		tx4, err := doubleSpend.Create(gen.GetAddressOutput(test_tx.Address4, grp.SendAmount2), address2Wallet)
		if err != nil {
			return jerr.Get("error saving tx4 to address 4", err)
		}
		jlog.Logf("tx4: %s\n", hs.GetTxString(tx4.GetHash()))
		tx5, err := doubleSpend.Create(gen.GetAddressOutput(test_tx.Address5, grp.SendAmount2), address3Wallet)
		if err != nil {
			return jerr.Get("error saving tx5 to address 5", err)
		}
		jlog.Logf("tx5: %s\n", hs.GetTxString(tx5.GetHash()))
		var address1Expected = grp.FundingValue - grp.SendAmount - memo.FeeP2pkh1In2OutTx
		var address2expected = grp.SendAmount - grp.SendAmount2 - memo.FeeP2pkh1In2OutTx
		if err := doubleSpend.CheckAddressBalance(test_tx.Address1String, address1Expected); err != nil {
			return jerr.Get("error address 1 balance does not match expected", err)
		}
		if err := doubleSpend.CheckAddressBalance(test_tx.Address2String, address2expected); err != nil {
			return jerr.Get("error address 2 balance does not match expected", err)
		}
		if err := doubleSpend.CheckAddressBalance(test_tx.Address3String, 0); err != nil {
			return jerr.Get("error address 3 balance does not match expected", err)
		}
		if err := doubleSpend.CheckAddressBalance(test_tx.Address4String, grp.SendAmount2); err != nil {
			return jerr.Get("error address 4 balance does not match expected", err)
		}
		if err := doubleSpend.CheckAddressBalance(test_tx.Address5String, 0); err != nil {
			return jerr.Get("error address 5 balance does not match expected", err)
		}
		if err := doubleSpend.SaveBlock([]*memo.Tx{tx3, tx5}); err != nil {
			return jerr.Get("error saving address 3 tx block", err)
		}
		if err := doubleSpend.CheckAddressBalance(test_tx.Address2String, 0); err != nil {
			return jerr.Get("error address 2 balance does not match expected after block", err)
		}
		if err := doubleSpend.CheckAddressBalance(test_tx.Address3String, address2expected); err != nil {
			return jerr.Get("error address 3 balance does not match expected after block", err)
		}
		if err := doubleSpend.CheckAddressBalance(test_tx.Address4String, 0); err != nil {
			return jerr.Get("error address 4 balance does not match expected after block", err)
		}
		if err := doubleSpend.CheckAddressBalance(test_tx.Address5String, grp.SendAmount2); err != nil {
			return jerr.Get("error address 5 balance does not match expected after block", err)
		}
		for i := 0; i < 5; i++ {
			if err := doubleSpend.SaveBlock(nil); err != nil {
				return jerr.Getf(err, "error saving address empty block %d", i)
			}
		}
		return nil
	},
}
