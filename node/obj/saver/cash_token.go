package saver

import (
	"context"
	"fmt"
	"log"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	item_cashtoken "github.com/memocash/index/db/item/cashtoken"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/tx/cashtoken"
	"github.com/memocash/index/ref/dbi"
)

// CashToken writes a cash_token_output row for every output whose lock
// script carries a well-formed CashTokens prefix. A malformed prefix is
// logged as a process error and skipped: the node would not have relayed
// such a tx, so seeing one means the decoder and the node disagree.
type CashToken struct {
	Verbose bool
}

func (c *CashToken) SaveTxs(ctx context.Context, block *dbi.Block) error {
	if block.IsNil() {
		return fmt.Errorf("error nil block")
	}
	var objects []db.Object
	for _, dbiTx := range block.Transactions {
		tx := dbiTx.MsgTx
		txHash := chainhash.Hash(dbiTx.Hash)
		for k := range tx.TxOut {
			if !cashtoken.HasPrefix(tx.TxOut[k].PkScript) {
				continue
			}
			token, _, err := cashtoken.Parse(tx.TxOut[k].PkScript)
			if err != nil {
				log.Printf("error parsing cash token prefix %s:%d; %v\n", txHash, k, err)
				continue
			}
			if c.Verbose {
				log.Printf("cash token output %s:%d category %s amount %d nft %t\n",
					txHash, k, chainhash.Hash(token.Category), token.Amount, token.HasNft)
			}
			objects = append(objects, item_cashtoken.NewOutput(txHash, uint32(k), *token))
		}
	}
	if err := db.Save(objects); err != nil {
		return fmt.Errorf("error saving cash token outputs; %w", err)
	}
	return nil
}

func NewCashToken(verbose bool) *CashToken {
	return &CashToken{
		Verbose: verbose,
	}
}
