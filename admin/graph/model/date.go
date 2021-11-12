package model

import (
	"encoding/json"
	"github.com/99designs/gqlgen/graphql"
	"github.com/jchavannes/jgo/jerr"
	"io"
	"time"
)

type Date time.Time

func MarshalDate(date Date) graphql.Marshaler {
	data, _ := json.Marshal(time.Time(date))
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, string(data))
	})
}

func UnmarshalDate(v interface{}) (Date, error) {
	switch v := v.(type) {
	case string:
		t, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			return Date{}, jerr.Get("error parsing date with rfc 3339 nano", err)
		}
		return Date(t), nil
	default:
		return Date{}, jerr.New("error unexpected hash index type not string")
	}
}
