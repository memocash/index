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

func (c *Client) updateDb(addresses []wallet.Addr) error {
	lastUpdates, err := c.Database.GetAddressLastUpdate(addresses)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error getting address last update; %w", err)
	}
	history, err := graph.GetHistory(c.GraphUrl, lastUpdates)
	if err != nil {
		return fmt.Errorf("error getting history txs; %w", err)
	}
	if err := c.Database.SaveTxs(history.GetAllTxs()); err != nil {
		return fmt.Errorf("error saving txs; %w", err)
	}
	var addressUpdates []graph.AddressUpdate
	for _, addrTxs := range history {
		var maxTime time.Time
		for _, tx := range addrTxs.Txs {
			if tx.Seen.After(maxTime) {
				maxTime = tx.Seen
			}
		}
		addressUpdates = append(addressUpdates, graph.AddressUpdate{
			Address: addrTxs.Address,
			Time:    maxTime,
		})
	}
	if err := c.Database.SetAddressLastUpdate(addressUpdates); err != nil {
		return fmt.Errorf("error setting address last updates for client update db; %w", err)
	}
	return nil
}

func (c *Client) Broadcast(txRaw string) error {
	if err := graph.Broadcast(c.GraphUrl, txRaw); err != nil {
		return fmt.Errorf("error broadcasting lib client tx; %w", err)
	}
	return nil
}

func (c *Client) GetBalance(addresses []wallet.Addr) (*Balance, error) {
	err := c.updateDb(addresses)
	if err != nil {
		return nil, fmt.Errorf("error updating db for get balance; %w", err)
	}
	balance, err := c.Database.GetAddressBalance(addresses)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting address balance from database; %w", err)
	}
	return balance, nil
}

func (c *Client) GetUtxos(addresses []wallet.Addr) ([]graph.Output, error) {
	err := c.updateDb(addresses)
	if err != nil {
		return nil, fmt.Errorf("error updating db for get utxos; %w", err)
	}
	utxos, err := c.Database.GetUtxos(addresses)
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
