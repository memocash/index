package load

import (
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
)

var TxInputs = dataloader.NewTxInputsLoader(dataloader.TxInputsLoaderConfig{
	Wait: defaultWait,
	Fetch: func(keys []string) ([][]*model.TxInput, []error) {
		var txHashes = make([][32]byte, len(keys))
		for i := range keys {
			txHash, err := chainhash.NewHashFromStr(keys[i])
			if err != nil {
				return nil, []error{jerr.Get("error getting tx hash for tx inputs dataloader", err)}
			}
			txHashes[i] = *txHash
		}
		txInputs, err := chain.GetTxInputsByHashes(txHashes)
		if err != nil {
			return nil, []error{jerr.Get("error getting tx inputs for model tx", err)}
		}
		var modelInputs = make([][]*model.TxInput, len(txHashes))
		for i := range txHashes {
			for _, txInput := range txInputs {
				if txHashes[i] == txInput.TxHash {
					modelInputs[i] = append(modelInputs[i], &model.TxInput{
						Hash:      chainhash.Hash(txInput.TxHash).String(),
						Index:     txInput.Index,
						PrevHash:  chainhash.Hash(txInput.PrevHash).String(),
						PrevIndex: txInput.PrevIndex,
						Script:    hex.EncodeToString(txInput.UnlockScript),
					})
				}
			}
		}
		return modelInputs, nil
	},
})
