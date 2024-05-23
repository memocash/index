package gen

import (
	"bytes"
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

func (c *Create) getMoreUTXOs() ([]memo.UTXO, error) {
	var moreUTXOs []memo.UTXO
	if !c.isEnoughSlpValue() {
		tokenHash := c.Request.GetTokenHash()
		if len(tokenHash) == 0 {
			return nil, fmt.Errorf("error unexpected empty token hash when need more slp value")
		}
		tokenUTXOs, err := c.Request.Getter.GetUTXOs(&memo.UTXORequest{TokenHash: tokenHash})
		if err != nil {
			return nil, fmt.Errorf("error getting more utxos for token; %w", err)
		}
		moreUTXOs = append(moreUTXOs, tokenUTXOs...)
	}
	enoughInputValue, err := c.isEnoughInputValue()
	if err != nil {
		return nil, fmt.Errorf("error determining if enough input value; %w", err)
	}
	if !enoughInputValue {
		satUTXOs, err := c.Request.Getter.GetUTXOs(nil)
		if err != nil {
			return nil, fmt.Errorf("error getting more utxos; %w", err)
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
		switch output.Script.(type) {
		case *script.P2pkh:
			fee += memo.OutputFeeP2PKH
			outputValues += output.Amount
		default:
			outputFee, err := output.GetValuePlusFee()
			if err != nil {
				return -1, fmt.Errorf("error getting memo output fee; %w", err)
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
		return false, fmt.Errorf("error getting min input; %w", err)
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
	return fmt.Errorf(
		"not enough value in inputs (minInput: %d, len(c.InputsToUse): %d, totalInputValue: %d):\n%s; %w",
		minInput, len(c.InputsToUse), c.getInputValue(), txInfo.GetString(), NotEnoughValueError)
}

func (c Create) GetQuickMemoTx() *memo.Tx {
	wireTx, _ := c.getWireTx()
	return GetMemoTx(wireTx, c.InputsToUse, c.Outputs)
}

func (c Create) getNotEnoughTokenValueError() error {
	tokenSendOutput := c.Request.GetTokenSendOutput()
	if tokenSendOutput == nil {
		return fmt.Errorf("unable to find token send output; %w", NotEnoughTokenValueError)
	}
	txInfo := parse.GetTxInfo(c.GetQuickMemoTx())
	return fmt.Errorf(
		"not enough token value in inputs (minTokenInput: %d, len(c.InputsToUse): %d, tokenInputTokenValue: %d):\n%s; %w",
		tokenSendOutput.GetTotalQuantity(), len(c.InputsToUse), c.getTokenInputValue(), txInfo.GetString(), NotEnoughTokenValueError)
}

func (c Create) getTxInputs() []memo.TxInput {
	var txInputs []memo.TxInput
	for _, input := range c.InputsToUse {
		txInputs = append(txInputs, input.Input)
	}
	return txInputs
}
