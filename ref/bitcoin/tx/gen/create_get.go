package gen

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

func (c *Create) getMoreUTXOs() ([]memo.UTXO, error) {
	var moreUTXOs []memo.UTXO
	if !c.isEnoughSlpValue() {
		tokenHash := c.Request.GetTokenHash()
		if len(tokenHash) == 0 {
			return nil, jerr.New("error unexpected empty token hash when need more slp value")
		}
		tokenUTXOs, err := c.Request.Getter.GetUTXOs(&memo.UTXORequest{TokenHash: tokenHash})
		if err != nil {
			return nil, jerr.Get("error getting more utxos for token", err)
		}
		moreUTXOs = append(moreUTXOs, tokenUTXOs...)
	}
	enoughInputValue, err := c.isEnoughInputValue()
	if err != nil {
		return nil, jerr.Get("error determining if enough input value", err)
	}
	if !enoughInputValue {
		satUTXOs, err := c.Request.Getter.GetUTXOs(nil)
		if err != nil {
			return nil, jerr.Get("error getting more utxos", err)
		}
		moreUTXOs = append(moreUTXOs, satUTXOs...)
	}
	return moreUTXOs, nil
}

func (c Create) getInputValue() int64 {
	var totalInputValue int64
	for _, input := range c.InputsToUse {
		totalInputValue += input.Input.Value
	}
	return totalInputValue
}

func (c Create) getMinInput() (int64, error) {
	var fee = memo.BaseTxFee + int64(len(c.InputsToUse))*memo.InputFeeP2PKH
	var outputValues int64
	for _, output := range c.Outputs {
		switch output.GetType() {
		case memo.OutputTypeP2PKH:
			fee += memo.OutputFeeP2PKH
			outputValues += output.Amount
		default:
			outputFee, err := output.GetValuePlusFee()
			if err != nil {
				return -1, jerr.Get("error getting memo output fee", err)
			}
			fee += outputFee
		}
	}
	var minInput = fee + outputValues
	return minInput, nil
}

func (c Create) getTokenInputValue() uint64 {
	tokenSendOutput := c.Request.GetTokenSendOutput()
	if tokenSendOutput == nil {
		return 0
	}
	var tokenInputValue uint64
	for _, inputToUse := range c.InputsToUse {
		if inputToUse.SlpToken != nil && bytes.Equal(inputToUse.SlpToken, tokenSendOutput.TokenHash) {
			tokenInputValue += inputToUse.SlpQuantity
		}
	}
	return tokenInputValue
}

func (c Create) getExpectedTokenValue() uint64 {
	tokenSendOutput := c.Request.GetTokenSendOutput()
	if tokenSendOutput == nil {
		return 0
	}
	return tokenSendOutput.GetTotalQuantity()
}

func (c Create) isComplete() bool {
	return c.isCorrectSlpValue() && c.isCorrectValue()
}

func (c Create) isCorrectValue() bool {
	minInput, err := c.getMinInput()
	if err != nil {
		return false
	}
	totalInputValue := c.getInputValue()
	return totalInputValue >= minInput && totalInputValue-minInput < memo.OutputFeeP2PKH
}

func (c Create) isCorrectSlpValue() bool {
	return c.getTokenInputValue() == c.getExpectedTokenValue()
}

func (c Create) isEnoughInputValue() (bool, error) {
	minInput, err := c.getMinInput()
	if err != nil {
		return false, jerr.Get("error getting min input", err)
	}
	totalInputValue := c.getInputValue()
	var additionInputForChange = memo.OutputFeeP2PKH + memo.DustMinimumOutput
	var hasEnoughInputValue = totalInputValue >= (minInput+additionInputForChange) ||
		(totalInputValue >= minInput && totalInputValue <= minInput+memo.OutputFeeP2PKH)
	return hasEnoughInputValue, nil
}

func (c Create) isEnoughSlpValue() bool {
	return c.getTokenInputValue() >= c.getExpectedTokenValue()
}

func (c Create) getNotEnoughValueError() error {
	minInput, _ := c.getMinInput()
	txInfo := parse.GetTxInfo(c.GetQuickMemoTx())
	return jerr.Getf(NotEnoughValueError,
		"not enough value in inputs (minInput: %d, len(c.InputsToUse): %d, totalInputValue: %d):\n%s",
		minInput, len(c.InputsToUse), c.getInputValue(), txInfo.GetString())
}

func (c Create) GetQuickMemoTx() *memo.Tx {
	wireTx, _ := c.getWireTx()
	return getMemoTx(wireTx, c.InputsToUse, c.Outputs)
}

func (c Create) getNotEnoughTokenValueError() error {
	tokenSendOutput := c.Request.GetTokenSendOutput()
	if tokenSendOutput == nil {
		return jerr.Get("unable to find token send output", NotEnoughTokenValueError)
	}
	txInfo := parse.GetTxInfo(c.GetQuickMemoTx())
	return jerr.Getf(NotEnoughTokenValueError,
		"not enough token value in inputs (minTokenInput: %d, len(c.InputsToUse): %d, tokenInputTokenValue: %d):\n%s",
		tokenSendOutput.GetTotalQuantity(), len(c.InputsToUse), c.getTokenInputValue(), txInfo.GetString())
}

func (c Create) getTxInputs() []memo.TxInput {
	var txInputs []memo.TxInput
	for _, input := range c.InputsToUse {
		txInputs = append(txInputs, input.Input)
	}
	return txInputs
}
