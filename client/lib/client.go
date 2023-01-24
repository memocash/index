package lib

import (
	"database/sql"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/client/lib/graph"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"time"
)

type Client struct {
	GraphUrl string
	Database Database
}

func (c *Client) updateDb(address *wallet.Addr) error {
	lastUpdate, err := c.Database.GetAddressLastUpdate(address)
	if err != nil && !jerr.HasError(err, sql.ErrNoRows.Error()) {
		return jerr.Get("error getting address last update", err)
	}
	txs, err := graph.GetHistory(c.GraphUrl, address, lastUpdate)
	if err != nil {
		return jerr.Get("error getting history txs", err)
	}
	if err := c.Database.SaveTxs(txs); err != nil {
		return jerr.Get("error saving txs", err)
	}
	var maxTime time.Time
	for _, tx := range txs {
		if tx.Seen.After(maxTime) {
			maxTime = tx.Seen
		}
	}
	if err := c.Database.SetAddressLastUpdate(address, maxTime); err != nil {
		return jerr.Get("error setting address last update for client update db", err)
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
