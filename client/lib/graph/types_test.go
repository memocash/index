package graph_test

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/memocash/index/client/lib/graph"
)

func TestSlpUnmarshalAmount(t *testing.T) {
	tests := []struct {
		json     string
		expected uint64
		wantErr  bool
	}{
		{`{"amount":"123"}`, 123, false},
		{`{"amount":123}`, 123, false},
		{`{"amount":"18446744073709551615"}`, math.MaxUint64, false},
		{`{"amount":18446744073709551615}`, math.MaxUint64, false},
		{`{"amount":null}`, 0, false},
		{`{}`, 0, false},
		{`{"amount":"abc"}`, 0, true},
		{`{"amount":"-1"}`, 0, true},
	}
	for _, test := range tests {
		var slp graph.Slp
		err := json.Unmarshal([]byte(test.json), &slp)
		if test.wantErr {
			if err == nil {
				t.Errorf("%s: expected error, got amount %d", test.json, slp.Amount)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error: %v", test.json, err)
			continue
		}
		if slp.Amount != test.expected {
			t.Errorf("%s: expected %d, got %d", test.json, test.expected, slp.Amount)
		}
	}
}

func TestOutputUnmarshalSlpFields(t *testing.T) {
	data := `{"index":1,"amount":546,"slp":{"hash":"abc","index":1,"token_hash":"def","amount":"1000","genesis":{"ticker":"TKN","decimals":2}}}`
	var output graph.Output
	if err := json.Unmarshal([]byte(data), &output); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Slp == nil {
		t.Fatal("expected slp to be set")
	}
	if output.Slp.Amount != 1000 || output.Slp.Index != 1 || output.Slp.TokenHash != "def" || output.Slp.Hash != "abc" {
		t.Errorf("unexpected slp fields: %+v", output.Slp)
	}
	if output.Slp.Genesis == nil || output.Slp.Genesis.Ticker != "TKN" || output.Slp.Genesis.Decimals != 2 {
		t.Errorf("unexpected slp genesis: %+v", output.Slp.Genesis)
	}
}
