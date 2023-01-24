package gen

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"sort"
)

func (c *Create) setInputs() error {
	hasEnoughInputValue, err := c.isEnoughInputValue()
	if err != nil {
		return jerr.Get("error determining if has enough input value", err)
	}
	if hasEnoughInputValue {
		return nil
	}
	for i := 0; i < len(c.PotentialInputs); i++ {
		var potentialInput = c.PotentialInputs[i]
		if potentialInput.IsSlp() || potentialInput.IsSellTokenInput() {
			continue
		}
		c.PotentialInputs = append(c.PotentialInputs[:i], c.PotentialInputs[i+1:]...)
		i--
		c.InputsToUse = append(c.InputsToUse, potentialInput)
		hasEnoughInputValue, err := c.isEnoughInputValue()
		if err != nil {
			return jerr.Get("error determining if has enough input value", err)
		}
		if hasEnoughInputValue {
			break
		}
	}
	return nil
}

func (c *Create) setChange() error {
	if ! c.isEnoughSlpValue() {
		// Don't set change until SLP is complete
		return nil
	}
	var totalInputValue = c.getInputValue()
	var inputAndBaseFee = memo.BaseTxFee + int64(len(c.InputsToUse))*memo.InputFeeP2PKH
	var outputValueRequired int64
	for _, spendOutput := range c.Outputs {
		outputValuePlusFee, err := spendOutput.GetValuePlusFee()
		if err != nil {
			return jerr.Get("error getting memo output fee", err)
		}
		outputValueRequired += outputValuePlusFee
	}
	var change = totalInputValue - inputAndBaseFee - outputValueRequired
	if change <= 0 {
		// No change yet
		return nil
	} else if change < memo.DustMinimumOutput+memo.OutputFeeP2PKH {
		// Not enough change for new output, add to existing output if possible
		if change < memo.OutputFeeP2PKH {
			for _, output := range c.Outputs {
				outputAddress := script.GetAddress(output.Script)
				if output.Amount > memo.DustMinimumOutput && outputAddress.IsSame(c.Request.Change.Main) {
					output.Amount += change
					break
				}
			}
		}
	} else {
		if ! c.Request.Change.Main.IsSet() {
			return jerr.New("change address not set")
		}
		change -= memo.OutputFeeP2PKH
		c.Outputs = append(c.Outputs, GetAddressOutput(c.Request.Change.Main, change))
	}
	return nil
}

func (c *Create) setSlpInputs() error {
	if c.isEnoughSlpValue() {
		return nil
	}
	tokenSendOutput := c.Request.GetTokenSendOutput()
	if tokenSendOutput == nil {
		return nil
	}
	currentInputsTokenValue := c.getTokenInputValue()
	var potentialTokenInputs []memo.UTXO
	for _, potentialInput := range c.PotentialInputs {
		if potentialInput.SlpToken == nil {
			continue
		}
		if bytes.Equal(potentialInput.SlpToken, tokenSendOutput.TokenHash) {
			potentialTokenInputs = append(potentialTokenInputs, potentialInput)
		}
	}
	sort.Slice(potentialTokenInputs, func(i, j int) bool {
		return potentialTokenInputs[i].SlpQuantity > potentialTokenInputs[j].SlpQuantity
	})

	// Check for exact match
	for _, potentialTokenInput := range potentialTokenInputs {
		if potentialTokenInput.SlpQuantity == tokenSendOutput.GetTotalQuantity()-currentInputsTokenValue {
			c.InputsToUse = append(c.InputsToUse, potentialTokenInput)
			break
		}
	}
loop:
	for i := 0; ! c.isEnoughSlpValue() && i < len(potentialTokenInputs); i++ {
		c.InputsToUse = append(c.InputsToUse, potentialTokenInputs[i])
		for j := range c.PotentialInputs {
			if bytes.Equal(c.PotentialInputs[j].Input.PrevOutHash, potentialTokenInputs[i].Input.PrevOutHash) &&
				c.PotentialInputs[j].Input.PrevOutIndex == potentialTokenInputs[i].Input.PrevOutIndex {
				c.PotentialInputs = append(c.PotentialInputs[:j], c.PotentialInputs[j+1:]...)
				continue loop
			}
		}
		return jerr.New("found unknown potential input")
	}
	return nil
}

func (c *Create) setSlpChange() error {
	tokenSendOutput := c.Request.GetTokenSendOutput()
	if tokenSendOutput == nil {
		return nil
	}
	if c.getTokenInputValue() > tokenSendOutput.GetTotalQuantity() {
		slpChange := c.Request.Change.GetSlp()
		if ! slpChange.IsSet() {
			return jerr.New("error slp change address not set")
		}
		c.Outputs = append(c.Outputs, GetAddressOutput(slpChange, memo.DustMinimumOutput))
		tokenSendOutput.Quantities = append(tokenSendOutput.Quantities, c.getTokenInputValue()-tokenSendOutput.GetTotalQuantity())
	}
	return nil
}
