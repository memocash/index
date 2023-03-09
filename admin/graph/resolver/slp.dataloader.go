package resolver

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
	"time"
)

var slpOutputLoaderConfig = dataloader.SlpOutputLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(keys []model.HashIndex) ([]*model.SlpOutput, []error) {
		var memoOuts = make([]memo.Out, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i].Hash)
			if err != nil {
				return nil, []error{jerr.Get("error parsing spend tx hash for output", err)}
			}
			memoOuts[i] = memo.Out{
				TxHash: hash[:],
				Index:  keys[i].Index,
			}
		}
		slpOutputs, err := slp.GetOutputs(memoOuts)
		if err != nil && !client.IsMessageNotSetError(err) {
			return nil, []error{jerr.Get("error getting slp output for slp genesis resolver", err)}
		}
		var modelSlpOutputs = make([]*model.SlpOutput, len(memoOuts))
		for i := range memoOuts {
			for _, slpOutput := range slpOutputs {
				if bytes.Equal(memoOuts[i].TxHash, slpOutput.TxHash[:]) && memoOuts[i].Index == slpOutput.Index {
					modelSlpOutputs[i] = &model.SlpOutput{
						Hash:      chainhash.Hash(slpOutput.TxHash).String(),
						Index:     slpOutput.Index,
						TokenHash: chainhash.Hash(slpOutput.TokenHash).String(),
						Amount:    slpOutput.Quantity,
					}
					break
				}
			}
		}
		return modelSlpOutputs, nil
	},
}

func SlpGenesisLoader(txHashStr string) (*model.SlpGenesis, error) {
	txHash, err := chainhash.NewHashFromStr(txHashStr)
	if err != nil {
		return nil, jerr.Get("error getting tx hash for slp genesis output resolver", err)
	}
	slpGenesis, err := slp.GetGenesis(*txHash)
	if err != nil {
		return nil, jerr.Get("error getting slp output for slp genesis resolver", err)
	}
	return &model.SlpGenesis{
		Hash:       chainhash.Hash(slpGenesis.TxHash).String(),
		TokenType:  model.Uint8(slpGenesis.TokenType),
		Decimals:   model.Uint8(slpGenesis.Decimals),
		BatonIndex: slpGenesis.BatonIndex,
		Ticker:     slpGenesis.Ticker,
		Name:       slpGenesis.Name,
		DocURL:     slpGenesis.DocUrl,
		DocHash:    hex.EncodeToString(slpGenesis.DocHash[:]),
	}, nil
}

func SlpBatonLoader(txHashStr string, index uint32) (*model.SlpBaton, error) {
	txHash, err := chainhash.NewHashFromStr(txHashStr)
	if err != nil {
		return nil, jerr.Get("error getting tx hash for slp genesis baton resolver", err)
	}
	slpBaton, err := slp.GetBaton(*txHash, index)
	if err != nil && !client.IsMessageNotSetError(err) {
		return nil, jerr.Get("error getting slp baton for slp genesis resolver", err)
	}
	if slpBaton == nil {
		return nil, nil
	}
	return &model.SlpBaton{
		Hash:      chainhash.Hash(slpBaton.TxHash).String(),
		Index:     slpBaton.Index,
		TokenHash: chainhash.Hash(slpBaton.TokenHash).String(),
	}, nil
}
