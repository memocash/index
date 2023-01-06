package lib

import (
	"database/sql"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/client/lib/graph"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Client struct {
	Database Database
}

func updateDb(address *wallet.Addr, c *Client) error {
	height, err := c.Database.GetAddressHeight(address)
	if err != nil && !jerr.HasError(err, sql.ErrNoRows.Error()) {
		return jerr.Get("error getting address height", err)
	}
	txs, err := graph.GetHistory(address, height)
	if err != nil {
		return jerr.Get("error getting history", err)
	}
	if err := c.Database.SaveTxs(txs); err != nil {
		return jerr.Get("error saving txs", err)
	}
	return nil
}

func (c *Client) GetBalance(address *wallet.Addr) (int64, error) {
	err := updateDb(address, c)
	if err != nil {
		return 0, jerr.Get("error updating db", err)
	}
	balance, err := c.Database.GetAddressBalance(address)
	if err != nil {
		if jerr.HasError(err, sql.ErrNoRows.Error()) {
			return 0, nil
		}
		return 0, jerr.Get("error saving outputs", err)
	}
	return balance, nil
}

func (c *Client) GetUtxos(address *wallet.Addr) ([]graph.Output, error) {
	err := updateDb(address, c)
	if err != nil {
		return nil, jerr.Get("error updating db", err)
	}
	utxos, err := c.Database.GetUtxos(address)
	if err != nil {
		if jerr.HasError(err, sql.ErrNoRows.Error()) {
			return nil, nil
		}
		return nil, jerr.Get("error saving outputs", err)
	}
	return utxos, nil
}

func NewClient(database Database) *Client {
	return &Client{
		Database: database,
	}
}
