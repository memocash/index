package slp

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"math/big"
)

type Status uint8

const (
	StatusPending Status = 0 // undecided; no verdict row is written
	StatusValid   Status = 1
	StatusInvalid Status = 2
	StatusSkip    Status = 3 // out of scope (COMMIT); never stored
)

type Reason uint8

const (
	ReasonNone            Reason = 0
	ReasonParse           Reason = 1 // strict parse failure
	ReasonTokenType       Reason = 2 // unhandled token type
	ReasonNotVoutZero     Reason = 3 // slp op_return not at vout 0
	ReasonNftChildGenesis Reason = 4 // nft child genesis constraints not met
	ReasonMintNoBaton     Reason = 5 // mint without valid baton input
	ReasonSendInputSum    Reason = 6 // send declared outputs exceed valid inputs
)

// ParentState is the resolved processing/validity state of an input's prevout.
type ParentState uint8

const (
	// ParentUnknown: parent's SLP standing not yet resolvable (tx not
	// ingested, or SLP transcription rows not yet written) - pending.
	ParentUnknown ParentState = iota
	// ParentNotSlp: parent definitively has no SLP row at this prevout
	// (vout 0 carries no SLP message, or transcription is complete without
	// one) - contributes nothing.
	ParentNotSlp
	// ParentPending: SLP row exists but the parent has no validity verdict yet.
	ParentPending
	ParentValid
	ParentInvalid
)

// Input is one tx input with its prevout's SLP resolution attached.
type Input struct {
	PrevHash [32]byte
	Index    uint32
	State    ParentState
	Output   *InputToken // SLP quantity output row at this prevout, if any
	Baton    *InputBaton // SLP baton row at this prevout, if any
}

// InputToken is an SLP quantity output spent by an input. TokenType is the
// type the parent chain declared for the token (resolved from the token's
// genesis); TypeKnown false means it could not be resolved yet, and any
// contribution this input would have made is treated as pending.
type InputToken struct {
	TokenHash [32]byte
	TokenType uint16
	TypeKnown bool
	Quantity  uint64
}

type InputBaton struct {
	TokenHash [32]byte
	TokenType uint16
	TypeKnown bool
}

// TxData is everything the pure validator needs about one SLP transaction.
type TxData struct {
	// Msg is the strict-parsed vout-0 message; nil means strict parse failed.
	Msg *Msg
	// Inputs in input order (input 0 matters for NFT child genesis).
	Inputs []Input
}

type Verdict struct {
	Status Status
	Reason Reason
}

var verdictValid = Verdict{Status: StatusValid}

var verdictPending = Verdict{Status: StatusPending}

func invalid(reason Reason) Verdict {
	return Verdict{Status: StatusInvalid, Reason: reason}
}

// Validate applies the SLP validity rules to a strict-parsed tx with resolved
// inputs. It is pure: verdicts depend only on TxData. The core invariant is
// fail-to-pending: whenever a relevant input's state is genuinely unknown the
// verdict is Pending, never a guess.
func Validate(data TxData) Verdict {
	if data.Msg == nil {
		return invalid(ReasonParse)
	}
	switch data.Msg.TokenType {
	case memo.SlpDefaultTokenType, memo.SlpNftGroupTokenType, memo.SlpNftChildTokenType:
	default:
		return invalid(ReasonTokenType)
	}
	switch data.Msg.TxType {
	case memo.SlpTxTypeGenesis:
		return validateGenesis(data)
	case memo.SlpTxTypeMint:
		return validateMint(data)
	case memo.SlpTxTypeSend:
		return validateSend(data)
	case memo.SlpTxTypeCommit:
		return Verdict{Status: StatusSkip}
	}
	return invalid(ReasonParse)
}

func validateGenesis(data TxData) Verdict {
	if data.Msg.TokenType != memo.SlpNftChildTokenType {
		// Type 1 and NFT group geneses are self-certifying
		return verdictValid
	}
	// NFT child genesis: input 0 must spend a valid NFT group (0x81) token
	// output with quantity > 0. (Declared constraints are parse rules.)
	if len(data.Inputs) == 0 {
		return invalid(ReasonNftChildGenesis)
	}
	var input = data.Inputs[0]
	switch input.State {
	case ParentUnknown, ParentPending:
		return verdictPending
	case ParentValid:
	default:
		return invalid(ReasonNftChildGenesis)
	}
	if input.Output == nil || input.Output.Quantity == 0 {
		return invalid(ReasonNftChildGenesis)
	}
	if !input.Output.TypeKnown {
		return verdictPending
	}
	if input.Output.TokenType != memo.SlpNftGroupTokenType {
		return invalid(ReasonNftChildGenesis)
	}
	return verdictValid
}

func validateMint(data TxData) Verdict {
	if data.Msg.TokenType == memo.SlpNftChildTokenType {
		// NFT children cannot be minted (also a parse rule)
		return invalid(ReasonMintNoBaton)
	}
	// Valid iff some input spends a valid baton of the declared token id and
	// token type; token and foreign/invalid inputs are ignored
	var anyPossible bool
	for _, input := range data.Inputs {
		switch input.State {
		case ParentUnknown:
			// Could be an unresolved baton of this token
			anyPossible = true
		case ParentPending:
			if input.Baton != nil && input.Baton.TokenHash == data.Msg.Mint.TokenHash &&
				(!input.Baton.TypeKnown || input.Baton.TokenType == data.Msg.TokenType) {
				anyPossible = true
			}
		case ParentValid:
			if input.Baton != nil && input.Baton.TokenHash == data.Msg.Mint.TokenHash {
				if !input.Baton.TypeKnown {
					anyPossible = true
				} else if input.Baton.TokenType == data.Msg.TokenType {
					return verdictValid
				}
			}
		}
	}
	if anyPossible {
		return verdictPending
	}
	return invalid(ReasonMintNoBaton)
}

func validateSend(data TxData) Verdict {
	var outputSum = new(big.Int)
	var quantity = new(big.Int)
	for _, q := range data.Msg.Send.Quantities {
		outputSum.Add(outputSum, quantity.SetUint64(q))
	}
	if outputSum.Sign() == 0 {
		// Zero-output sends are self-evidently valid
		return verdictValid
	}
	// Valid iff the declared output sum does not exceed the sum of valid
	// inputs of the same token id and token type; invalid and foreign inputs
	// contribute zero and never invalidate
	var validSum = new(big.Int)
	var pendingSum = new(big.Int)
	var anyUnknown bool
	for _, input := range data.Inputs {
		switch input.State {
		case ParentUnknown:
			anyUnknown = true
		case ParentPending:
			if input.Output != nil && input.Output.TokenHash == data.Msg.Send.TokenHash &&
				(!input.Output.TypeKnown || input.Output.TokenType == data.Msg.TokenType) {
				pendingSum.Add(pendingSum, quantity.SetUint64(input.Output.Quantity))
			}
		case ParentValid:
			if input.Output != nil && input.Output.TokenHash == data.Msg.Send.TokenHash {
				if !input.Output.TypeKnown {
					pendingSum.Add(pendingSum, quantity.SetUint64(input.Output.Quantity))
				} else if input.Output.TokenType == data.Msg.TokenType {
					validSum.Add(validSum, quantity.SetUint64(input.Output.Quantity))
				}
			}
		}
	}
	if validSum.Cmp(outputSum) >= 0 {
		return verdictValid
	}
	if anyUnknown {
		return verdictPending
	}
	if validSum.Add(validSum, pendingSum).Cmp(outputSum) < 0 {
		// Even if every pending input resolves valid the outputs are unfunded
		return invalid(ReasonSendInputSum)
	}
	return verdictPending
}
