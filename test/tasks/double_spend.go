package tasks

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/jchavannes/jgo/jlog"
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
		balanceAddress1, err := doubleSpend.GetAddressBalance(test_tx.Address1String)
		if err != nil {
			return jerr.Get("error getting balance for address 1", err)
		}
		if balanceAddress1 > grp.FundingValue {
			return jerr.Newf("error address 1 balance greater than funding amount: %s %s",
				jfmt.AddCommas(balanceAddress1), jfmt.AddCommas(grp.FundingValue))
		}
		balanceAddress2, err := doubleSpend.GetAddressBalance(test_tx.Address2String)
		if err != nil {
			return jerr.Get("error getting balance for address 2", err)
		}
		if balanceAddress2 != grp.SendAmount {
			return jerr.Newf("error address 2 balance not equal send amount: %s", jfmt.AddCommas(balanceAddress2))
		}
		balanceAddress3, err := doubleSpend.GetAddressBalance(test_tx.Address3String)
		if err != nil {
			return jerr.Get("error getting balance for address 3", err)
		}
		if balanceAddress3 != 0 {
			return jerr.Newf("error address 3 balance not equal 0: %s", jfmt.AddCommas(balanceAddress3))
		}
		if err := doubleSpend.SaveBlock(tx3); err != nil {
			return jerr.Get("error saving address 3 tx block", err)
		}
		balanceAddress2, err = doubleSpend.GetAddressBalance(test_tx.Address2String)
		if err != nil {
			return jerr.Get("error getting balance for address 2", err)
		}
		if balanceAddress2 != 0 {
			return jerr.Newf("error address 2 balance not equal 0: %s", jfmt.AddCommas(balanceAddress2))
		}
		balanceAddress3, err = doubleSpend.GetAddressBalance(test_tx.Address3String)
		if err != nil {
			return jerr.Get("error getting balance for address 3", err)
		}
		if balanceAddress3 != grp.SendAmount {
			return jerr.Newf("error address 3 balance not equal send amount: %s", jfmt.AddCommas(balanceAddress3))
		}
		return nil
	},
}
