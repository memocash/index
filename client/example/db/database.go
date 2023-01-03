package db

import (
	"database/sql"
	"github.com/jchavannes/jgo/jerr"
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

func NewDatabase() (*Database, error) {
	db, err := sql.Open("sqlite3", "./example.db")
	if err != nil {
		return nil, jerr.Get("error opening database", err)
	}
	if err := initDb(db); err != nil {
		return nil, jerr.Get("error initializing database", err)
	}
	return &Database{
		Db: db,
	}, nil
}

func (d *Database) GetAddressBalance(address *wallet.Addr) (int64, error) {
	query := "" +
		"SELECT " +
		"   outputs.address, " +
		"   COUNT(DISTINCT (outputs.hash || outputs.`index`)) AS output_count, " +
		"   IFNULL(SUM(CASE WHEN inputs.hash IS NULL THEN 1 ELSE 0 END), 0) AS utxo_count, " +
		"   IFNULL(SUM(CASE WHEN inputs.hash IS NULL THEN outputs.value ELSE 0 END), 0) AS balance " +
		"FROM outputs " +
		"LEFT JOIN inputs ON (inputs.prev_hash = outputs.hash AND inputs.prev_index = outputs.`index`) " +
		"WHERE outputs.address = ? " +
		"GROUP BY outputs.address "
	var result struct {
		Address string
		Balance int64
	}
	if err := d.Db.QueryRow(query, address.String()).Scan(&result); err != nil {
		return 0, jerr.Get("error getting address balance exec query", err)
	}
	return result.Balance, nil
}

func (d *Database) GetAddressHeight(address *wallet.Addr) (int64, error) {
	query := "" +
		"SELECT " +
		"   outputs.address, " +
		"   MAX(blocks.height) AS height " +
		"FROM outputs " +
		"JOIN block_txs ON (block_txs.tx_hash = outputs.hash) " +
		"JOIN blocks on (blocks.hash = block_txs.block_hash) " +
		"WHERE outputs.address = ? " +
		"GROUP BY outputs.address "
	var result struct {
		Address string
		Height  int64
	}
	if err := d.Db.QueryRow(query, address.String()).Scan(&result); err != nil {
		return 0, jerr.Get("error getting address balance exec query", err)
	}
	return result.Height, nil
}

func (d * Database) GetUtxos(address *wallet.Addr) ([]graph.Output, error){
	query := "" +
		"SELECT outputs.* FROM outputs " +
		"LEFT JOIN inputs ON (inputs.prev_hash = outputs.hash AND inputs.prev_index = outputs.`index`) " +
		"WHERE outputs.address = ?" +
		"AND inputs.hash IS NULL"
	rows, err := d.Db.Query(query, address.String())
	if err != nil {
		return nil, jerr.Get("error getting address utxos select query", err)
	}
	var results []graph.Output
	for rows.Next() {
		var result graph.Output
		if err := rows.Scan(&result); err != nil {
			return nil, jerr.Get("error getting address utxos scan query", err)
		}
		results = append(results, result)
	}
	return results, nil
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
				Query:     "INSERT OR IGNORE INTO blocks (hash, timestamp, height) VALUES (?, ?)",
				Variables: []interface{}{block.Hash, block.Timestamp.Format(time.RFC3339Nano), block.Height},
			}, Query{
				Name:      "block_txs",
				Query:     "INSERT OR IGNORE INTO block_txs (block_hash, tx_hash) VALUES (?, ?)",
				Variables: []interface{}{block.Hash, tx.Hash},
			})
		}
		if err := execQueries(d.Db, queries); err != nil {
			return jerr.Get("error saving txs", err)
		}
	}
	return nil
}
