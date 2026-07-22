package memo

import (
	"testing"
)

func TestNormalizePageLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit uint32
		want  uint32
	}{
		{name: "default", want: DefaultPageSize},
		{name: "explicit", limit: 25, want: 25},
		{name: "above default", limit: DefaultPageSize + 1, want: DefaultPageSize + 1},
		{name: "maximum", limit: MaxPageSize, want: MaxPageSize},
		{name: "above maximum", limit: MaxPageSize + 1, want: MaxPageSize},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NormalizePageLimit(test.limit); got != test.want {
				t.Fatalf("NormalizePageLimit(%d) = %d, want %d", test.limit, got, test.want)
			}
		})
	}
}
