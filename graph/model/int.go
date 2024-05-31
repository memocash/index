package model

import (
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
	"strconv"
)

type Uint8 uint8

func MarshalUint8(ui Uint8) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.FormatUint(uint64(ui), 10))
	})
}

func UnmarshalUint8(v interface{}) (Uint8, error) {
	switch v := v.(type) {
	case string:
		u64, err := strconv.ParseUint(v, 10, 8)
		return Uint8(u64), err
	case int:
		return Uint8(v), nil
	case int64:
		return Uint8(v), nil
	case uint:
		return Uint8(v), nil
	case uint8:
		return Uint8(v), nil
	case uint64:
		return Uint8(v), nil
	case json.Number:
		u64, err := strconv.ParseUint(string(v), 10, 8)
		return Uint8(u64), err
	default:
		return 0, fmt.Errorf("%T is not an uint", v)
	}
}

type Uint32 uint32

func UnmarshalUint32(v interface{}) (Uint32, error) {
	switch v := v.(type) {
	case string:
		u64, err := strconv.ParseUint(v, 10, 32)
		return Uint32(u64), err
	case int:
		return Uint32(v), nil
	case int64:
		return Uint32(v), nil
	case uint:
		return Uint32(v), nil
	case uint8:
		return Uint32(v), nil
	case uint64:
		return Uint32(v), nil
	case json.Number:
		u64, err := strconv.ParseUint(string(v), 10, 32)
		return Uint32(u64), err
	default:
		return 0, fmt.Errorf("%T is not an uint32", v)
	}
}

func IntPtr(i int) *int {
	return &i
}
