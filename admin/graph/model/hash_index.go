package model

import (
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
)

type HashIndex struct {
	Hash  string
	Index uint32
}

func MarshalHashIndex(hashIndex HashIndex) graphql.Marshaler {
	data, _ := json.Marshal(hashIndex)
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, string(data))
	})
}

func UnmarshalHashIndex(v interface{}) (HashIndex, error) {
	switch v := v.(type) {
	case string:
		var hashIndex HashIndex
		if err := json.Unmarshal([]byte(v), &hashIndex); err != nil {
			return HashIndex{}, fmt.Errorf("error unmarshalling hash index: %s", err)
		}
		return hashIndex, nil
	default:
		return HashIndex{}, fmt.Errorf("error unexpected hash index type not string")
	}
}
