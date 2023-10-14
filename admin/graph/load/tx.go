package load

import (
	"context"
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/node/act/tx_raw"
	"time"
)

var TxSeen = dataloader.NewTxSeenLoader(dataloader.TxSeenLoaderConfig{
	Wait: defaultWait,
	Fetch: func(keys []string) ([]*model.Date, []error) {
		var txHashes = make([][32]byte, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i])
			if err != nil {
				return nil, []error{jerr.Get("error parsing spend tx hash for output", err)}
			}
			txHashes[i] = *hash
		}
		txSeens, err := chain.GetTxSeens(txHashes)
		if err != nil && !client.IsResourceUnavailableError(err) {
			return nil, []error{jerr.Get("error getting tx seens", err)}
		}
		var modelTxSeens = make([]*model.Date, len(txHashes))
		for i := range txHashes {
			for _, txSeen := range txSeens {
				if txSeen.TxHash == txHashes[i] {
					if modelTxSeens[i] == nil || time.Time(*modelTxSeens[i]).After(txSeen.Timestamp) {
						var modelDate = model.Date(txSeen.Timestamp)
						modelTxSeens[i] = &modelDate
					}
				}
			}
			if modelTxSeens[i] == nil {
				return nil, []error{jerr.Newf("tx seen not found for hash: %s", txHashes[i])}
			}
		}
		return modelTxSeens, nil
	},
})

var TxRaw = dataloader.NewTxRawLoader(dataloader.TxRawLoaderConfig{
	Wait: defaultWait,
	Fetch: func(keys []string) ([]*model.Tx, []error) {
		var txHashes = make([][32]byte, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i])
			if err != nil {
				return nil, []error{jerr.Get("error parsing tx hash for raw loader", err)}
			}
			txHashes[i] = *hash
		}
		txRaws, err := tx_raw.Get(txHashes)
		if err != nil {
			return nil, []error{jerr.Get("error getting tx raws for tx dataloader", err)}
		}
		txsWithRaw := make([]*model.Tx, len(txHashes))
		for h := range txHashes {
			for r := range txRaws {
				if txRaws[r].Hash == txHashes[h] {
					txsWithRaw[h] = &model.Tx{
						Hash: chainhash.Hash(txRaws[r].Hash).String(),
						Raw:  hex.EncodeToString(txRaws[r].Raw),
					}
					break
				}
			}
			if txsWithRaw[h] == nil {
				return nil, []error{jerr.Newf("tx raw not found for hash: %s", chainhash.Hash(txHashes[h]))}
			}
		}
		return txsWithRaw, nil
	},
})

func Tx(ctx context.Context, txHash string) (*model.Tx, error) {
	var tx = &model.Tx{Hash: txHash}
	if HasField(ctx, "raw") {
		txWithRaw, err := TxRaw.Load(txHash)
		if err != nil {
			return nil, jerr.Get("error getting tx raw from dataloader for post resolver", err)
		}
		tx.Raw = txWithRaw.Raw
	}
	if HasField(ctx, "seen") {
		txSeen, err := TxSeen.Load(txHash)
		if err != nil {
			return nil, jerr.Get("error getting tx seen for tx loader", err)
		}
		tx.Seen = *txSeen
	}
	return tx, nil
}
