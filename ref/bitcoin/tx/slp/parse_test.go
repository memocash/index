package slp_test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/slp"
)

type scriptTest struct {
	Msg    string `json:"msg"`
	Script string `json:"script"`
	Code   *int   `json:"code"`
}

// TestHasSlpLokad covers scripts too short or malformed to hold a lokad push,
// including a bare OP_RETURN, which must not panic.
func TestHasSlpLokad(t *testing.T) {
	var lokadScript = append([]byte{0x6a, byte(len(memo.PrefixSlp))}, memo.PrefixSlp...)
	for _, test := range []struct {
		name     string
		pkScript []byte
		want     bool
	}{
		{"empty", nil, false},
		{"bare op_return", []byte{0x6a}, false},
		{"op_return truncated push", []byte{0x6a, 0x04, 0x53}, false},
		{"not op_return", []byte{0x76}, false},
		{"slp lokad", lokadScript, true},
	} {
		if got := slp.HasSlpLokad(test.pkScript); got != test.want {
			t.Errorf("%s: HasSlpLokad = %v, want %v", test.name, got, test.want)
		}
	}
}

// TestScriptVectors runs the strict parser against the canonical
// slp-unit-test-data script vectors (github.com/simpleledger/slp-unit-test-data).
// A nil code means the script must parse; code 255 means it must parse but
// carry an unsupported token type (rejected by the validator, not the parser);
// any other code means the parser must reject it.
func TestScriptVectors(t *testing.T) {
	data, err := os.ReadFile("testdata/script_tests.json")
	if err != nil {
		t.Fatalf("read script vectors: %v", err)
	}
	var tests []scriptTest
	if err := json.Unmarshal(data, &tests); err != nil {
		t.Fatalf("unmarshal script vectors: %v", err)
	}
	if len(tests) == 0 {
		t.Fatal("no script vectors")
	}
	for _, test := range tests {
		pkScript, err := hex.DecodeString(test.Script)
		if err != nil {
			t.Fatalf("decode script for %q: %v", test.Msg, err)
		}
		msg, parseErr := slp.Parse(pkScript)
		switch {
		case test.Code == nil:
			if parseErr != nil {
				t.Errorf("%q: expected parse ok, got: %v", test.Msg, parseErr)
			}
		case *test.Code == 255:
			if parseErr != nil {
				t.Errorf("%q: expected parse ok with unsupported token type, got: %v", test.Msg, parseErr)
			} else if msg.TokenType == memo.SlpDefaultTokenType ||
				msg.TokenType == memo.SlpNftGroupTokenType ||
				msg.TokenType == memo.SlpNftChildTokenType {
				t.Errorf("%q: expected unsupported token type, got 0x%x", test.Msg, msg.TokenType)
			}
		default:
			if parseErr == nil {
				t.Errorf("%q: expected parse error (code %d), got ok", test.Msg, *test.Code)
			}
		}
	}
}
