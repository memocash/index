package model

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
)

type Bytes []byte

func MarshalBytes(bytes Bytes) graphql.Marshaler {
	data, _ := json.Marshal(hex.EncodeToString(bytes))
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, string(data))
	})
}

func UnmarshalBytes(v interface{}) (Bytes, error) {
	switch v := v.(type) {
	case string:
		bytes, err := hex.DecodeString(v)
		if err != nil {
			return Bytes{}, fmt.Errorf("error unmarshal parsing bytes as byte slice; %w", err)
		}
		return bytes, nil
	default:
		return Bytes{}, fmt.Errorf("error unmarshal unexpected bytes type not string: %T", v)
	}
}
