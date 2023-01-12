package lib

import (
	"database/sql"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/client/lib/graph"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Client struct {
	GraphUrl string
	Database Database
}

func (c *Client) updateDb(address *wallet.Addr) error {
	height, err := c.Database.GetAddressHeight(address)
	if err != nil && !jerr.HasError(err, sql.ErrNoRows.Error()) {
		return jerr.Get("error getting address height", err)
	}
	history, err := graph.GetHistory(c.GraphUrl, address, height)
	if err != nil {
		return jerr.Get("error getting history", err)
	}
	for _, txs := range [][]graph.Tx{history.Outputs, history.Spends} {
		if err := c.Database.SaveTxs(txs); err != nil {
			return jerr.Get("error saving txs", err)
		}
	}
	var maxHeightOutput int64
	for _, output := range history.Outputs {
		if len(output.Blocks) > 0 && output.Blocks[0].Height > maxHeightOutput {
			maxHeightOutput = output.Blocks[0].Height
		}
	}
	var maxHeightSpends int64
	for _, spend := range history.Spends {
		if len(spend.Blocks) > 0 && spend.Blocks[0].Height > maxHeightSpends {
			maxHeightSpends = spend.Blocks[0].Height
		}
	}
	var maxHeight int64
	if maxHeightOutput > maxHeightSpends {
		maxHeight = maxHeightSpends
	} else {
		maxHeight = maxHeightOutput
	}
	if err := c.Database.SetAddressHeight(address, maxHeight); err != nil {
		return jerr.Get("error setting address height for client update db", err)
	}
	return nil
}

func (c *Client) GetBalance(address *wallet.Addr) (int64, error) {
	err := c.updateDb(address)
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
	err := c.updateDb(address)
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

func NewClient(graphUrl string, database Database) *Client {
	return &Client{
		GraphUrl: graphUrl,
		Database: database,
	}
}
