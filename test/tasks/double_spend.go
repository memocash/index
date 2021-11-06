package tasks

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/tx/hs"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/server/test/grp"
	"github.com/memocash/server/test/suite"
)

var doubleSpendTest = suite.Test{
	Name: TestDoubleSpend,
	Test: func(r *suite.TestRequest) error {
		var doubleSpend = &grp.DoubleSpend{}
		if err := doubleSpend.Init(); err != nil {
			return jerr.Get("error initializing double spend group", err)
		}
		tx2, err := doubleSpend.Create(gen.GetAddressOutput(test_tx.Address2, grp.SendAmount))
		if err != nil {
			return jerr.Get("error saving tx1 to address 2", err)
		}
		jlog.Logf("tx2: %s\n", hs.GetTxString(tx2.GetHash()))
		tx3, err := doubleSpend.Create(gen.GetAddressOutput(test_tx.Address3, grp.SendAmount))
		if err != nil {
			return jerr.Get("error saving tx2 to address 3", err)
		}
		jlog.Logf("tx3: %s\n", hs.GetTxString(tx3.GetHash()))
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
