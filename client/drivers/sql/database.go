package sql

import (
	"database/sql"
	"fmt"
	"github.com/memocash/index/client/lib/graph"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"time"
)

type Query struct {
	Name      string
	Query     string
	Variables []interface{}
}

type Database struct {
	Db *sql.DB
}

func NewDatabase(db *sql.DB) (*Database, error) {
	if err := initDb(db); err != nil {
		return nil, fmt.Errorf("%w; error initializing database", err)
	}
	return &Database{
		Db: db,
	}, nil
}

func (d *Database) GetAddressBalance(address *wallet.Addr) (int64, error) {
	query := "" +
		"SELECT " +
		"   outputs.address, " +
		"   IFNULL(SUM(CASE WHEN inputs.hash IS NULL THEN outputs.value ELSE 0 END), 0) AS balance " +
		"FROM outputs " +
		"LEFT JOIN inputs ON (inputs.prev_hash = outputs.hash AND inputs.prev_index = outputs.`index`) " +
		"WHERE outputs.address = ? " +
		"GROUP BY outputs.address "
	var result struct {
		Address string
		Balance int64
	}
	if err := d.Db.QueryRow(query, address.String()).Scan(&result.Address, &result.Balance); err != nil {
		return 0, fmt.Errorf("%w; error getting address balance exec query", err)
	}
	return result.Balance, nil
}

func (d *Database) GetAddressLastUpdate(address *wallet.Addr) (time.Time, error) {
	query := "" +
		"SELECT time " +
		"FROM address_updates " +
		"WHERE address = ? "
	var result struct {
		Time int64 `db:"time"`
	}
	if err := d.Db.QueryRow(query, address.String()).Scan(&result.Time); err != nil {
		return time.Time{}, fmt.Errorf("%w; error address last update exec query", err)
	}
	return time.Unix(result.Time, 0), nil
}

func (d *Database) GetUtxos(address *wallet.Addr) ([]graph.Output, error) {
	query := "" +
		"SELECT outputs.* FROM outputs " +
		"LEFT JOIN inputs ON (inputs.prev_hash = outputs.hash AND inputs.prev_index = outputs.`index`) " +
		"WHERE outputs.address = ?" +
		"AND inputs.hash IS NULL"
	rows, err := d.Db.Query(query, address.String())
	if err != nil {
		return nil, fmt.Errorf("%w; error getting address utxos select query", err)
	}
	var results []graph.Output
	for rows.Next() {
		var result graph.Output
		if err := rows.Scan(&result.Hash, &result.Index, &result.Lock.Address, &result.Amount); err != nil {
			return nil, fmt.Errorf("%w; error getting address utxos scan query", err)
		}
		results = append(results, result)
	}
	return results, nil
}

func (d *Database) SetAddressLastUpdate(address *wallet.Addr, t time.Time) error {
	if t.Unix() <= 0 {
		return nil
	}
	query := "INSERT OR REPLACE INTO address_updates (address, time) VALUES (?, ?)"
	if _, err := d.Db.Exec(query, address.String(), t.Unix()); err != nil {
		return fmt.Errorf("%w; error updating address last update", err)
	}
	return nil
}

func (d *Database) SaveTxs(txs []graph.Tx) error {
	for _, tx := range txs {
		var queries = []Query{{
			Name:      "txs",
			Query:     "INSERT OR IGNORE INTO txs (hash) VALUES (?)",
			Variables: []interface{}{tx.Hash},
		}}
		for _, input := range tx.Inputs {
			queries = append(queries, Query{
				Name:      "inputs",
				Query:     "INSERT OR IGNORE INTO inputs (hash, `index`, prev_hash, prev_index) VALUES (?, ?, ?, ?)",
				Variables: []interface{}{tx.Hash, input.Index, input.PrevHash, input.PrevIndex},
			})
		}
		for _, output := range tx.Outputs {
			queries = append(queries, Query{
				Name:      "outputs",
				Query:     "INSERT OR IGNORE INTO outputs (hash, `index`, address, value) VALUES (?, ?, ?, ?)",
				Variables: []interface{}{tx.Hash, output.Index, output.Lock.Address, output.Amount},
			})
		}
		for _, block := range tx.Blocks {
			queries = append(queries, Query{
				Name:      "blocks",
				Query:     "INSERT OR IGNORE INTO blocks (hash, timestamp, height) VALUES (?, ?, ?)",
				Variables: []interface{}{block.Hash, block.Timestamp.Format(time.RFC3339Nano), block.Height},
			}, Query{
				Name:      "block_txs",
				Query:     "INSERT OR IGNORE INTO block_txs (block_hash, tx_hash) VALUES (?, ?)",
				Variables: []interface{}{block.Hash, tx.Hash},
			})
		}
		if err := execQueries(d.Db, queries); err != nil {
			return fmt.Errorf("%w; error saving txs", err)
		}
	}
	return nil
}
