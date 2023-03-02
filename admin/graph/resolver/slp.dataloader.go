package resolver

import (
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/slp"
)

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
		Decimals:   model.Uint8(slpGenesis.Decimals),
		BatonIndex: slpGenesis.BatonIndex,
		Ticker:     slpGenesis.Ticker,
		Name:       slpGenesis.Name,
		DocURL:     slpGenesis.DocUrl,
		DocHash:    hex.EncodeToString(slpGenesis.DocHash[:]),
	}, nil
}

func SlpOutputLoader(txHashStr string, index uint32) (*model.SlpOutput, error) {
	txHash, err := chainhash.NewHashFromStr(txHashStr)
	if err != nil {
		return nil, jerr.Get("error getting tx hash for slp genesis output resolver", err)
	}
	slpOutput, err := slp.GetOutput(*txHash, index)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return nil, jerr.Get("error getting slp output for slp genesis resolver", err)
	}
	if slpOutput == nil {
		return nil, nil
	}
	return &model.SlpOutput{
		Hash:   chainhash.Hash(slpOutput.TxHash).String(),
		Index:  slpOutput.Index,
		Amount: slpOutput.Quantity,
	}, nil
}

func SlpBatonLoader(txHashStr string, index uint32) (*model.SlpBaton, error) {
	txHash, err := chainhash.NewHashFromStr(txHashStr)
	if err != nil {
		return nil, jerr.Get("error getting tx hash for slp genesis baton resolver", err)
	}
	slpBaton, err := slp.GetBaton(*txHash, index)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return nil, jerr.Get("error getting slp baton for slp genesis resolver", err)
	}
	if slpBaton == nil {
		return nil, nil
	}
	return &model.SlpBaton{
		Hash:  chainhash.Hash(slpBaton.TxHash).String(),
		Index: slpBaton.Index,
	}, nil
}
