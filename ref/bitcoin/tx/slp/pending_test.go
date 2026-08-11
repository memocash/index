package slp_test

import (
	"testing"

	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
)

var testTokenHash = [32]byte{1, 2, 3}

func sendMsg(tokenType uint16, quantities ...uint64) *slp.Msg {
	return &slp.Msg{
		TokenType: tokenType,
		TxType:    memo.SlpTxTypeSend,
		Send:      &slp.SendMsg{TokenHash: testTokenHash, Quantities: quantities},
	}
}

func tokenInput(state slp.ParentState, quantity uint64) slp.Input {
	return slp.Input{
		State: state,
		Output: &slp.InputToken{
			TokenHash: testTokenHash,
			TokenType: memo.SlpDefaultTokenType,
			TypeKnown: true,
			Quantity:  quantity,
		},
	}
}

// TestOutOfOrderResolution walks a send's input through the states it passes
// when a child arrives before its parent (mempool order, CTOR in-block order,
// or split across concurrent shard batches): the verdict must be pending at
// every unresolved step and valid at the end - never invalid.
func TestOutOfOrderResolution(t *testing.T) {
	var msg = sendMsg(memo.SlpDefaultTokenType, 100)
	var steps = []struct {
		name  string
		input slp.Input
		want  slp.Status
	}{
		{"parent not ingested", slp.Input{State: slp.ParentUnknown}, slp.StatusPending},
		{"parent transcribed, no verdict", tokenInput(slp.ParentPending, 100), slp.StatusPending},
		{"parent row present, type unresolved", func() slp.Input {
			input := tokenInput(slp.ParentValid, 100)
			input.Output.TypeKnown = false
			return input
		}(), slp.StatusPending},
		{"parent valid", tokenInput(slp.ParentValid, 100), slp.StatusValid},
	}
	for _, step := range steps {
		verdict := slp.Validate(slp.TxData{Msg: msg, Inputs: []slp.Input{step.input}})
		if verdict.Status != step.want {
			t.Errorf("%s: expected status %d, got %d (reason %d)",
				step.name, step.want, verdict.Status, verdict.Reason)
		}
		if verdict.Status == slp.StatusInvalid {
			t.Errorf("%s: out-of-order arrival must never yield invalid", step.name)
		}
	}
}

// TestSendEarlyDecisions covers the bound-based decisions: enough valid input
// decides valid without waiting on pending inputs, and an optimistic bound
// that still falls short decides invalid without waiting.
func TestSendEarlyDecisions(t *testing.T) {
	var msg = sendMsg(memo.SlpDefaultTokenType, 100)
	verdict := slp.Validate(slp.TxData{Msg: msg, Inputs: []slp.Input{
		tokenInput(slp.ParentValid, 100),
		tokenInput(slp.ParentPending, 50),
	}})
	if verdict.Status != slp.StatusValid {
		t.Errorf("valid inputs already cover outputs: expected valid, got %d", verdict.Status)
	}
	verdict = slp.Validate(slp.TxData{Msg: msg, Inputs: []slp.Input{
		tokenInput(slp.ParentValid, 10),
		tokenInput(slp.ParentPending, 50),
	}})
	if verdict.Status != slp.StatusInvalid || verdict.Reason != slp.ReasonSendInputSum {
		t.Errorf("outputs unfunded even optimistically: expected invalid input sum, got %d (reason %d)",
			verdict.Status, verdict.Reason)
	}
	verdict = slp.Validate(slp.TxData{Msg: msg, Inputs: []slp.Input{
		tokenInput(slp.ParentValid, 10),
		tokenInput(slp.ParentPending, 90),
	}})
	if verdict.Status != slp.StatusPending {
		t.Errorf("pending input decides outcome: expected pending, got %d", verdict.Status)
	}
}

// TestSendOverflow exercises sums beyond 2^64-1 on both sides; a uint64
// accumulator would wrap and misjudge both directions.
func TestSendOverflow(t *testing.T) {
	var half = uint64(0x8000000000000000)
	var max = uint64(0xFFFFFFFFFFFFFFFF)
	// Outputs sum to 1.5*2^64, inputs to ~2*2^64: valid
	verdict := slp.Validate(slp.TxData{
		Msg: sendMsg(memo.SlpDefaultTokenType, half, half, half),
		Inputs: []slp.Input{
			tokenInput(slp.ParentValid, max),
			tokenInput(slp.ParentValid, max),
		},
	})
	if verdict.Status != slp.StatusValid {
		t.Errorf("overflowing outputs covered by overflowing inputs: expected valid, got %d (reason %d)",
			verdict.Status, verdict.Reason)
	}
	// Outputs sum to ~2*2^64 (wrapping to ~0 in uint64), inputs far short: invalid
	verdict = slp.Validate(slp.TxData{
		Msg:    sendMsg(memo.SlpDefaultTokenType, max, max),
		Inputs: []slp.Input{tokenInput(slp.ParentValid, 1)},
	})
	if verdict.Status != slp.StatusInvalid {
		t.Errorf("overflowing outputs unfunded: expected invalid, got %d", verdict.Status)
	}
}

// TestMintBatonResolution covers the mint pending/decided transitions.
func TestMintBatonResolution(t *testing.T) {
	var msg = &slp.Msg{
		TokenType: memo.SlpDefaultTokenType,
		TxType:    memo.SlpTxTypeMint,
		Mint:      &slp.MintMsg{TokenHash: testTokenHash, Quantity: 5},
	}
	var baton = func(state slp.ParentState, typeKnown bool) slp.Input {
		return slp.Input{
			State: state,
			Baton: &slp.InputBaton{
				TokenHash: testTokenHash,
				TokenType: memo.SlpDefaultTokenType,
				TypeKnown: typeKnown,
			},
		}
	}
	var cases = []struct {
		name  string
		input slp.Input
		want  slp.Status
	}{
		{"unknown parent could be a baton", slp.Input{State: slp.ParentUnknown}, slp.StatusPending},
		{"baton row present, parent undecided", baton(slp.ParentPending, true), slp.StatusPending},
		{"baton row present, type unresolved", baton(slp.ParentValid, false), slp.StatusPending},
		{"valid baton", baton(slp.ParentValid, true), slp.StatusValid},
		{"invalid baton parent", baton(slp.ParentInvalid, true), slp.StatusInvalid},
		{"no baton anywhere", tokenInput(slp.ParentValid, 5), slp.StatusInvalid},
	}
	for _, c := range cases {
		verdict := slp.Validate(slp.TxData{Msg: msg, Inputs: []slp.Input{c.input}})
		if verdict.Status != c.want {
			t.Errorf("%s: expected status %d, got %d (reason %d)", c.name, c.want, verdict.Status, verdict.Reason)
		}
	}
}
