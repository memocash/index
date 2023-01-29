package lib

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/memocash/index/client/lib/graph"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"time"
)

type Client struct {
	GraphUrl string
	Database Database
}

func (c *Client) updateDb(address wallet.Addr) error {
	lastUpdate, err := c.Database.GetAddressLastUpdate(address)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w; error getting address last update", err)
	}
	txs, err := graph.GetHistory(c.GraphUrl, address, lastUpdate)
	if err != nil {
		return fmt.Errorf("%w; error getting history txs", err)
	}
	if err := c.Database.SaveTxs(txs); err != nil {
		return fmt.Errorf("%w; error saving txs", err)
	}
	var maxTime time.Time
	for _, tx := range txs {
		if tx.Seen.After(maxTime) {
			maxTime = tx.Seen
		}
	}
	if err := c.Database.SetAddressLastUpdate(address, maxTime); err != nil {
		return fmt.Errorf("%w; error setting address last update for client update db", err)
	}
	return nil
}

func (c *Client) GetBalance(address wallet.Addr) (int64, error) {
	err := c.updateDb(address)
	if err != nil {
		return 0, fmt.Errorf("%w; error updating db", err)
	}
	balance, err := c.Database.GetAddressBalance(address)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("%w; error saving outputs", err)
	}
	return balance, nil
}

func (c *Client) GetUtxos(address wallet.Addr) ([]graph.Output, error) {
	err := c.updateDb(address)
	if err != nil {
		return nil, fmt.Errorf("%w; error updating db", err)
	}
	utxos, err := c.Database.GetUtxos(address)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w; error saving outputs", err)
	}
	return utxos, nil
}

func NewClient(graphUrl string, database Database) *Client {
	return &Client{
		GraphUrl: graphUrl,
		Database: database,
	}
}
