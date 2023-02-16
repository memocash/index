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
		return fmt.Errorf("error getting address last update; %w", err)
	}
	txs, err := graph.GetHistory(c.GraphUrl, address, lastUpdate)
	if err != nil {
		return fmt.Errorf("error getting history txs; %w", err)
	}
	if err := c.Database.SaveTxs(txs); err != nil {
		return fmt.Errorf("error saving txs; %w", err)
	}
	var maxTime time.Time
	for _, tx := range txs {
		if tx.Seen.After(maxTime) {
			maxTime = tx.Seen
		}
	}
	if err := c.Database.SetAddressLastUpdate(address, maxTime); err != nil {
		return fmt.Errorf("error setting address last update for client update db; %w", err)
	}
	return nil
}

func (c *Client) Broadcast(txRaw string) error {
	if err := graph.Broadcast(c.GraphUrl, txRaw); err != nil {
		return fmt.Errorf("error broadcasting lib client tx; %w", err)
	}
	return nil
}

func (c *Client) GetBalance(address wallet.Addr) (int64, error) {
	err := c.updateDb(address)
	if err != nil {
		return 0, fmt.Errorf("error updating db for get balance; %w", err)
	}
	balance, err := c.Database.GetAddressBalance(address)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting address balance from database; %w", err)
	}
	return balance, nil
}

func (c *Client) GetUtxos(address wallet.Addr) ([]graph.Output, error) {
	err := c.updateDb(address)
	if err != nil {
		return nil, fmt.Errorf("error updating db for get utxos; %w", err)
	}
	utxos, err := c.Database.GetUtxos(address)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting utxos from database; %w", err)
	}
	return utxos, nil
}

func NewClient(graphUrl string, database Database) *Client {
	return &Client{
		GraphUrl: graphUrl,
		Database: database,
	}
}
