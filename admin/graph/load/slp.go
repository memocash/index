package load

import (
	"bytes"
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/ref/bitcoin/memo"
)

var SlpOutput = dataloader.NewSlpOutputLoader(dataloader.SlpOutputLoaderConfig{
	Wait: defaultWait,
	Fetch: func(keys []model.HashIndex) ([]*model.SlpOutput, []error) {
		var memoOuts = make([]memo.Out, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i].Hash)
			if err != nil {
				return nil, []error{jerr.Get("error parsing tx hash for slp output dataloader", err)}
			}
			memoOuts[i] = memo.Out{
				TxHash: hash[:],
				Index:  keys[i].Index,
			}
		}
		slpOutputs, err := slp.GetOutputs(memoOuts)
		if err != nil && !client.IsMessageNotSetError(err) {
			return nil, []error{jerr.Get("error getting slp output from dataloader", err)}
		}
		var modelSlpOutputs = make([]*model.SlpOutput, len(memoOuts))
		for i := range memoOuts {
			for _, slpOutput := range slpOutputs {
				if bytes.Equal(memoOuts[i].TxHash, slpOutput.TxHash[:]) && memoOuts[i].Index == slpOutput.Index {
					modelSlpOutputs[i] = &model.SlpOutput{
						Hash:      slpOutput.TxHash,
						Index:     slpOutput.Index,
						TokenHash: slpOutput.TokenHash,
						Amount:    slpOutput.Quantity,
					}
					break
				}
			}
		}
		return modelSlpOutputs, nil
	},
})

var SlpGenesis = dataloader.NewSlpGenesisLoader(dataloader.SlpGenesisLoaderConfig{
	Wait: defaultWait,
	Fetch: func(keys []string) ([]*model.SlpGenesis, []error) {
		var txHashes = make([][32]byte, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i])
			if err != nil {
				return nil, []error{jerr.Get("error parsing tx hash for slp baton dataloader", err)}
			}
			txHashes[i] = *hash
		}
		slpGeneses, err := slp.GetGeneses(txHashes)
		if err != nil && !client.IsMessageNotSetError(err) {
			return nil, []error{jerr.Get("error getting slp geneses from dataloader", err)}
		}
		var modelSlpGeneses = make([]*model.SlpGenesis, len(txHashes))
		for i := range txHashes {
			for _, slpGenesis := range slpGeneses {
				if txHashes[i] == slpGenesis.TxHash {
					modelSlpGeneses[i] = &model.SlpGenesis{
						Hash:       chainhash.Hash(slpGenesis.TxHash).String(),
						TokenType:  model.Uint8(slpGenesis.TokenType),
						Decimals:   model.Uint8(slpGenesis.Decimals),
						BatonIndex: slpGenesis.BatonIndex,
						Ticker:     slpGenesis.Ticker,
						Name:       slpGenesis.Name,
						DocURL:     slpGenesis.DocUrl,
						DocHash:    hex.EncodeToString(slpGenesis.DocHash[:]),
					}
					break
				}
			}
		}
		return modelSlpGeneses, nil
	},
})

var SlpBaton = dataloader.NewSlpBatonLoader(dataloader.SlpBatonLoaderConfig{
	Wait: defaultWait,
	Fetch: func(keys []model.HashIndex) ([]*model.SlpBaton, []error) {
		var memoOuts = make([]memo.Out, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i].Hash)
			if err != nil {
				return nil, []error{jerr.Get("error parsing tx hash for slp baton dataloader", err)}
			}
			memoOuts[i] = memo.Out{
				TxHash: hash[:],
				Index:  keys[i].Index,
			}
		}
		slpBatons, err := slp.GetBatons(memoOuts)
		if err != nil && !client.IsMessageNotSetError(err) {
			return nil, []error{jerr.Get("error getting slp batons from dataloader", err)}
		}
		var modelSlpBatons = make([]*model.SlpBaton, len(memoOuts))
		for i := range memoOuts {
			for _, slpBaton := range slpBatons {
				if bytes.Equal(memoOuts[i].TxHash, slpBaton.TxHash[:]) && memoOuts[i].Index == slpBaton.Index {
					modelSlpBatons[i] = &model.SlpBaton{
						Hash:      chainhash.Hash(slpBaton.TxHash).String(),
						Index:     slpBaton.Index,
						TokenHash: chainhash.Hash(slpBaton.TokenHash).String(),
					}
					break
				}
			}
		}
		return modelSlpBatons, nil
	},
})
