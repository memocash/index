package sql

import (
	"database/sql"
	"fmt"
	"github.com/memocash/index/client/lib"
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
	Prefix string
	Db     *sql.DB
}

func (d *Database) GetTableName(table string) string {
	t, ok := tables[table]
	if !ok {
		return ""
	}
	return t.GetName(d.Prefix)
}

func (d *Database) GetInsert(table string, values map[string]interface{}) *Query {
	t, ok := tables[table]
	if !ok {
		return nil
	}
	var q = t.GetInsert(d.Prefix, values)
	return &q
}

func NewDatabase(db *sql.DB, prefix string) (*Database, error) {
	if err := initDb(db, prefix); err != nil {
		return nil, fmt.Errorf("error initializing database; %w", err)
	}
	return &Database{
		Db:     db,
		Prefix: prefix,
	}, nil
}

func (d *Database) GetAddressBalance(address wallet.Addr) (*lib.Balance, error) {
	query := "" +
		"SELECT " +
		"   IFNULL(SUM(CASE WHEN inputs.hash IS NULL THEN outputs.value ELSE 0 END), 0) AS balance, " +
		"   IFNULL(SUM(CASE WHEN inputs.hash IS NULL THEN 1 ELSE 0 END), 0) AS utxo_count, " +
		"   IFNULL(SUM(CASE WHEN inputs.hash IS NULL AND slp_outputs.hash IS NULL AND slp_batons.hash IS NULL THEN outputs.value ELSE 0 END), 0) AS spendable, " +
		"   IFNULL(SUM(CASE WHEN inputs.hash IS NULL AND slp_outputs.hash IS NULL AND slp_batons.hash IS NULL THEN 1 ELSE 0 END), 0) AS spendable_count " +
		"FROM " + d.GetTableName(TableOutputs) + " outputs " +
		"LEFT JOIN " + d.GetTableName(TableInputs) + " inputs ON (inputs.prev_hash = outputs.hash AND inputs.prev_index = outputs.`index`) " +
		"LEFT JOIN " + d.GetTableName(TableSlpOutputs) + " slp_outputs ON (slp_outputs.hash = outputs.hash AND slp_outputs.`index` = outputs.`index`) " +
		"LEFT JOIN " + d.GetTableName(TableSlpBatons) + " slp_batons ON (slp_batons.hash = outputs.hash AND slp_batons.`index` = outputs.`index`) " +
		"WHERE outputs.address = ? " +
		"GROUP BY outputs.address "
	var result = new(lib.Balance)
	if err := d.Db.QueryRow(query, address.String()).Scan(
		&result.Balance,
		&result.UtxoCount,
		&result.Spendable,
		&result.SpendableCount,
	); err != nil {
		return nil, fmt.Errorf("error getting address balance exec query; %w", err)
	}
	return result, nil
}

func (d *Database) GetAddressLastUpdate(address wallet.Addr) (time.Time, error) {
	query := "" +
		"SELECT time " +
		"FROM " + d.GetTableName(TableAddressUpdates) + " " +
		"WHERE address = ? "
	var result struct {
		Time int64 `db:"time"`
	}
	if err := d.Db.QueryRow(query, address.String()).Scan(&result.Time); err != nil {
		return time.Time{}, fmt.Errorf("error address last update exec query; %w", err)
	}
	return time.Unix(result.Time, 0), nil
}

func (d *Database) GetUtxos(address wallet.Addr) ([]graph.Output, error) {
	query := "" +
		"SELECT " +
		"	outputs.hash, " +
		"	outputs.`index`, " +
		"	outputs.address, " +
		"	outputs.value " +
		"FROM " + d.GetTableName(TableOutputs) + " outputs " +
		"LEFT JOIN " + d.GetTableName(TableInputs) + " inputs ON (inputs.prev_hash = outputs.hash AND inputs.prev_index = outputs.`index`) " +
		"WHERE outputs.address = ? " +
		"AND inputs.hash IS NULL"
	rows, err := d.Db.Query(query, address.String())
	if err != nil {
		return nil, fmt.Errorf("error getting address utxos select query; %w", err)
	}
	var results []graph.Output
	for rows.Next() {
		var result graph.Output
		if err := rows.Scan(&result.Hash, &result.Index, &result.Lock.Address, &result.Amount); err != nil {
			return nil, fmt.Errorf("error getting address utxos scan query; %w", err)
		}
		results = append(results, result)
	}
	return results, nil
}

func (d *Database) SetAddressLastUpdate(address wallet.Addr, t time.Time) error {
	if t.Unix() <= 0 {
		return nil
	}
	query := d.GetInsert(TableAddressUpdates, map[string]interface{}{
		"address": address.String(),
		"time":    t.Unix(),
	})
	if err := execQueries(d.Db, []*Query{query}); err != nil {
		return fmt.Errorf("error updating address last update; %w", err)
	}
	return nil
}

func (d *Database) SaveTxs(txs []graph.Tx) error {
	for _, tx := range txs {
		var queries = []*Query{d.GetInsert(TableTxs, map[string]interface{}{"hash": tx.Hash})}
		for _, input := range tx.Inputs {
			queries = append(queries,
				d.GetInsert(TableInputs, map[string]interface{}{
					"hash":       tx.Hash,
					"index":      input.Index,
					"prev_hash":  input.PrevHash,
					"prev_index": input.PrevIndex,
				}))
		}
		for _, output := range tx.Outputs {
			queries = append(queries,
				d.GetInsert(TableOutputs, map[string]interface{}{
					"hash":    tx.Hash,
					"index":   output.Index,
					"address": output.Lock.Address,
					"value":   output.Amount,
				}))
			if output.Slp != nil {
				queries = append(queries,
					d.GetInsert(TableSlpOutputs, map[string]interface{}{
						"hash":       tx.Hash,
						"index":      output.Index,
						"token_hash": output.Slp.TokenHash,
						"amount":     output.Slp.Amount,
					}))
				if output.Slp.Genesis != nil {
					queries = append(queries,
						d.GetInsert(TableSlpGeneses, map[string]interface{}{
							"hash":       output.Slp.Genesis.Hash,
							"token_type": output.Slp.Genesis.TokenType,
							"decimals":   output.Slp.Genesis.Decimals,
							"ticker":     output.Slp.Genesis.Ticker,
							"name":       output.Slp.Genesis.Name,
							"doc_url":    output.Slp.Genesis.DocUrl,
						}))
				}
			}
			if output.SlpBaton != nil {
				queries = append(queries,
					d.GetInsert(TableSlpBatons, map[string]interface{}{
						"hash":       tx.Hash,
						"index":      output.Index,
						"token_hash": output.SlpBaton.TokenHash,
					}))
				if output.SlpBaton.Genesis != nil {
					queries = append(queries,
						d.GetInsert(TableSlpGeneses, map[string]interface{}{
							"hash":       output.SlpBaton.Genesis.Hash,
							"token_type": output.SlpBaton.Genesis.TokenType,
							"decimals":   output.SlpBaton.Genesis.Decimals,
							"ticker":     output.SlpBaton.Genesis.Ticker,
							"name":       output.SlpBaton.Genesis.Name,
							"doc_url":    output.SlpBaton.Genesis.DocUrl,
						}))
				}
			}
		}
		for _, block := range tx.Blocks {
			queries = append(queries,
				d.GetInsert(TableBlocks, map[string]interface{}{
					"hash":      block.Hash,
					"timestamp": block.Timestamp.Format(time.RFC3339Nano),
					"height":    block.Height,
				}),
				d.GetInsert(TableBlockTxs, map[string]interface{}{
					"block_hash": block.Hash,
					"tx_hash":    tx.Hash,
				}))
		}
		if err := execQueries(d.Db, queries); err != nil {
			return fmt.Errorf("error saving txs; %w", err)
		}
	}
	return nil
}
