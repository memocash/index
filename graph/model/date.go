package model

import (
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
	"time"
)

type Date time.Time

func MarshalDate(date Date) graphql.Marshaler {
	data, _ := json.Marshal(time.Time(date).Format(time.RFC3339Nano))
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, string(data))
	})
}

func UnmarshalDate(v interface{}) (Date, error) {
	switch v := v.(type) {
	case string:
		t, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			return Date{}, fmt.Errorf("error parsing date with rfc 3339 nano; %w", err)
		}
		return Date(t), nil
	default:
		return Date{}, fmt.Errorf("error unexpected date type not string: %T", v)
	}
}
