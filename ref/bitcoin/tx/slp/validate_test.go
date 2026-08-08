package slp_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
)

type inputTestTx struct {
	Tx    string `json:"tx"`
	Valid bool   `json:"valid"`
}

type inputTest struct {
	Description       string        `json:"description"`
	When              []inputTestTx `json:"when"`
	Should            []inputTestTx `json:"should"`
	AllowInconclusive bool          `json:"allow_inconclusive"`
}

// whenTx is a context tx with its given validity and derived SLP rows.
type whenTx struct {
	msg     *slp.Msg
	valid   bool
	outputs map[uint32]slp.InputToken
	batons  map[uint32]slp.InputBaton
}

func decodeTx(t *testing.T, raw string) (*wire.MsgTx, [32]byte) {
	rawBytes, err := hex.DecodeString(raw)
	if err != nil {
		t.Fatalf("decode tx hex: %v", err)
	}
	var msgTx wire.MsgTx
	if err := msgTx.Deserialize(bytes.NewReader(rawBytes)); err != nil {
		t.Fatalf("deserialize tx: %v", err)
	}
	return &msgTx, msgTx.TxHash()
}

// deriveRows mirrors the transcription layer: token quantity outputs and
// batons declared by a tx's message, limited to outputs that exist, with zero
// quantities skipped.
func deriveRows(msgTx *wire.MsgTx, txHash [32]byte, msg *slp.Msg) (map[uint32]slp.InputToken, map[uint32]slp.InputBaton) {
	var outputs = make(map[uint32]slp.InputToken)
	var batons = make(map[uint32]slp.InputBaton)
	if msg == nil {
		return outputs, batons
	}
	var outputCount = uint32(len(msgTx.TxOut))
	var addOutput = func(index uint32, tokenHash [32]byte, quantity uint64) {
		if quantity > 0 && index < outputCount {
			outputs[index] = slp.InputToken{
				TokenHash: tokenHash,
				TokenType: msg.TokenType,
				TypeKnown: true,
				Quantity:  quantity,
			}
		}
	}
	var addBaton = func(index int, tokenHash [32]byte) {
		if index > 0 && uint32(index) < outputCount {
			batons[uint32(index)] = slp.InputBaton{
				TokenHash: tokenHash,
				TokenType: msg.TokenType,
				TypeKnown: true,
			}
		}
	}
	switch msg.TxType {
	case memo.SlpTxTypeGenesis:
		addOutput(memo.SlpMintTokenIndex, txHash, msg.Genesis.Quantity)
		addBaton(msg.Genesis.BatonVout, txHash)
	case memo.SlpTxTypeMint:
		addOutput(memo.SlpMintTokenIndex, msg.Mint.TokenHash, msg.Mint.Quantity)
		addBaton(msg.Mint.BatonVout, msg.Mint.TokenHash)
	case memo.SlpTxTypeSend:
		for i, quantity := range msg.Send.Quantities {
			addOutput(uint32(i+1), msg.Send.TokenHash, quantity)
		}
	}
	return outputs, batons
}

func parseVout0(msgTx *wire.MsgTx) *slp.Msg {
	if len(msgTx.TxOut) == 0 {
		return nil
	}
	msg, err := slp.Parse(msgTx.TxOut[0].PkScript)
	if err != nil {
		return nil
	}
	return msg
}

// buildTxData resolves a should-tx against the closed universe of when-txs:
// parents outside the universe are affirmatively not SLP, token ids outside
// the universe are affirmatively not geneses.
func buildTxData(msgTx *wire.MsgTx, whenTxs map[[32]byte]*whenTx) slp.TxData {
	var data = slp.TxData{Msg: parseVout0(msgTx)}
	for _, txIn := range msgTx.TxIn {
		var input = slp.Input{
			PrevHash: txIn.PreviousOutPoint.Hash,
			Index:    txIn.PreviousOutPoint.Index,
			State:    slp.ParentNotSlp,
		}
		if parent, ok := whenTxs[input.PrevHash]; ok {
			if output, ok := parent.outputs[input.Index]; ok {
				input.Output = &output
			}
			if baton, ok := parent.batons[input.Index]; ok {
				input.Baton = &baton
			}
			if input.Output != nil || input.Baton != nil {
				if parent.valid {
					input.State = slp.ParentValid
				} else {
					input.State = slp.ParentInvalid
				}
			}
		}
		data.Inputs = append(data.Inputs, input)
	}
	return data
}

// TestTxInputVectors runs the validator against the canonical
// slp-unit-test-data input vectors: each case gives context txs with assumed
// validity and txs to judge.
func TestTxInputVectors(t *testing.T) {
	data, err := os.ReadFile("testdata/tx_input_tests.json")
	if err != nil {
		t.Fatalf("read input vectors: %v", err)
	}
	var tests []inputTest
	if err := json.Unmarshal(data, &tests); err != nil {
		t.Fatalf("unmarshal input vectors: %v", err)
	}
	if len(tests) == 0 {
		t.Fatal("no input vectors")
	}
	for _, test := range tests {
		var whenTxs = make(map[[32]byte]*whenTx)
		for _, when := range test.When {
			msgTx, txHash := decodeTx(t, when.Tx)
			msg := parseVout0(msgTx)
			outputs, batons := deriveRows(msgTx, txHash, msg)
			whenTxs[txHash] = &whenTx{msg: msg, valid: when.Valid, outputs: outputs, batons: batons}
		}
		for _, should := range test.Should {
			msgTx, txHash := decodeTx(t, should.Tx)
			verdict := slp.Validate(buildTxData(msgTx, whenTxs))
			var expected = slp.StatusInvalid
			if should.Valid {
				expected = slp.StatusValid
			}
			if verdict.Status == expected {
				continue
			}
			if test.AllowInconclusive && verdict.Status == slp.StatusPending {
				continue
			}
			t.Errorf("%q tx %x: expected %d, got %d (reason %d)",
				test.Description, txHash[:4], expected, verdict.Status, verdict.Reason)
		}
	}
}
