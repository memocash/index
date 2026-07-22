package attach

import "testing"

func TestUnmarshalBooleanDefault(t *testing.T) {
	tests := []struct {
		name      string
		arguments map[string]interface{}
		want      bool
	}{
		{name: "absent", arguments: map[string]interface{}{}, want: true},
		{name: "null", arguments: map[string]interface{}{"newest": nil}, want: true},
		{name: "explicit true", arguments: map[string]interface{}{"newest": true}, want: true},
		{name: "explicit false", arguments: map[string]interface{}{"newest": false}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := unmarshalBooleanDefault(test.arguments, "newest", true); got != test.want {
				t.Fatalf("unmarshalBooleanDefault() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestUnmarshalPageLimit(t *testing.T) {
	limit, err := unmarshalPageLimit(map[string]interface{}{"limit": "25"}, "test field", 100)
	if err != nil || limit != 25 {
		t.Fatalf("unmarshalPageLimit() = (%d, %v), want (25, nil)", limit, err)
	}
	if _, err := unmarshalPageLimit(map[string]interface{}{"limit": "101"}, "test field", 100); err == nil {
		t.Fatal("unmarshalPageLimit() error = nil, want over-limit error")
	}
	if _, err := unmarshalPageLimit(map[string]interface{}{"limit": "invalid"}, "test field", 100); err == nil {
		t.Fatal("unmarshalPageLimit() error = nil, want parse error")
	}
}
