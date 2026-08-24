package model_test

import (
	"bytes"
	"encoding/json"
	"math"
	"testing"

	"github.com/memocash/index/graph/model"
)

// Uint64 must marshal as a quoted JSON string: JavaScript clients parse
// unquoted numbers as float64 and lose digits above 2^53-1, while SLP
// quantities use the full uint64 range.
func TestMarshalUint64String(t *testing.T) {
	tests := []struct {
		value    model.Uint64
		expected string
	}{
		{0, `"0"`},
		{46300, `"46300"`},
		{math.MaxInt64, `"9223372036854775807"`},
		{math.MaxUint64, `"18446744073709551615"`},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		model.MarshalUint64(test.value).MarshalGQL(&buf)
		if buf.String() != test.expected {
			t.Errorf("marshal %d = %s, expected %s", uint64(test.value), buf.String(), test.expected)
		}
		var roundTrip string
		if err := json.Unmarshal(buf.Bytes(), &roundTrip); err != nil {
			t.Errorf("marshal %d output %s is not a JSON string; %v", uint64(test.value), buf.String(), err)
		}
	}
}

func TestUnmarshalUint64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected model.Uint64
	}{
		{"18446744073709551615", model.Uint64(math.MaxUint64)},
		{json.Number("18446744073709551615"), model.Uint64(math.MaxUint64)},
		{int64(46300), 46300},
		{uint64(math.MaxUint64), model.Uint64(math.MaxUint64)},
	}
	for _, test := range tests {
		value, err := model.UnmarshalUint64(test.input)
		if err != nil {
			t.Errorf("error unmarshalling %v (%T); %v", test.input, test.input, err)
		} else if value != test.expected {
			t.Errorf("unmarshal %v (%T) = %d, expected %d", test.input, test.input, uint64(value), uint64(test.expected))
		}
	}
	if _, err := model.UnmarshalUint64("not a number"); err == nil {
		t.Errorf("expected error unmarshalling non-numeric string")
	}
}
