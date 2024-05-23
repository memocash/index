package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type BitcomRequest struct {
	Wallet   Wallet
	Filename string
	Filetype string
	Contents []byte
}

func Bitcom(request BitcomRequest) (*memo.Tx, error) {
	tx, err := SimpleSingle(request.Wallet, []*memo.Output{{
		Script: &script.Save{
			Filename: request.Filename,
			Filetype: request.Filetype,
			Contents: request.Contents,
		},
	}})
	if err != nil {
		return nil, fmt.Errorf("error creating bitcom tx; %w", err)
	}
	return tx, nil
}
