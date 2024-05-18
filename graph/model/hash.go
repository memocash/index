package model

import (
	"encoding/json"
	"github.com/99designs/gqlgen/graphql"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"io"
)

type Hash [32]byte

func (h Hash) String() string {
	return chainhash.Hash(h).String()
}

func MarshalHash(hash Hash) graphql.Marshaler {
	data, _ := json.Marshal(chainhash.Hash(hash).String())
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, string(data))
	})
}

func UnmarshalHash(v interface{}) (Hash, error) {
	switch v := v.(type) {
	case string:
		hash, err := chainhash.NewHashFromStr(v)
		if err != nil {
			return Hash{}, jerr.Get("error unmarshal parsing hash as chainhash", err)
		}
		return Hash(*hash), nil
	default:
		return Hash{}, jerr.Newf("error unmarshal unexpected hash type not string: %T", v)
	}
}

func HashesToArrays(hashes []Hash) [][32]byte {
	var hashArrays [][32]byte
	for _, hash := range hashes {
		hashArrays = append(hashArrays, hash)
	}
	return hashArrays
}
