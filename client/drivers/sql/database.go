package sql

import (
	"database/sql"
	"fmt"
	"github.com/jchavannes/jgo/db_util"
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

func (d *Database) GetAddressBalance(addresses []wallet.Addr) (*lib.Balance, error) {
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
		"WHERE outputs.address IN (" + db_util.GetQuestionMarksCombined(len(addresses)) + ")"
	var addressStrings = make([]interface{}, len(addresses))
	for i := range addresses {
		addressStrings[i] = addresses[i].String()
	}
	var result = new(lib.Balance)
	if err := d.Db.QueryRow(query, addressStrings...).Scan(
		&result.Balance,
		&result.UtxoCount,
		&result.Spendable,
		&result.SpendableCount,
	); err != nil {
		return nil, fmt.Errorf("error getting address balance exec query; %w", err)
	}
	return result, nil
}

func (d *Database) GetAddressLastUpdate(addresses []wallet.Addr) ([]graph.AddressUpdate, error) {
	query := "" +
		"SELECT address, time " +
		"FROM " + d.GetTableName(TableAddressUpdates) + " " +
		"WHERE address IN (" + db_util.GetQuestionMarksCombined(len(addresses)) + ") " +
		"GROUP BY address"
	var addressStrings = make([]interface{}, len(addresses))
	for i := range addresses {
		addressStrings[i] = addresses[i].String()
	}
	rows, err := d.Db.Query(query, addressStrings...)
	if err != nil {
		return nil, fmt.Errorf("error address last update exec query; %w", err)
	}
	var addressUpdates []graph.AddressUpdate
	for rows.Next() {
		var result struct {
			Address string `db:"address"`
			Time    int64  `db:"time"`
		}
		if err := rows.Scan(&result.Address, &result.Time); err != nil {
			return nil, fmt.Errorf("error address last update scan query; %w", err)
		}
		addr, err := wallet.GetAddrFromString(result.Address)
		if err != nil {
			return nil, fmt.Errorf("error getting address from string; %w", err)
		}
		addressUpdates = append(addressUpdates, graph.AddressUpdate{
			Address: *addr,
			Time:    time.Unix(result.Time, 0),
		})
	}
AddressLoop:
	for _, address := range addresses {
		for _, addressUpdate := range addressUpdates {
			if addressUpdate.Address.String() == address.String() {
				continue AddressLoop
			}
		}
		addressUpdates = append(addressUpdates, graph.AddressUpdate{
			Address: address,
			Time:    time.Unix(0, 0),
		})
	}
	return addressUpdates, nil
}

func (d *Database) GetUtxos(addresses []wallet.Addr) ([]graph.Output, error) {
	query := "" +
		"SELECT " +
		"	outputs.hash, " +
		"	outputs.`index`, " +
		"	outputs.address, " +
		"	outputs.value " +
		"FROM " + d.GetTableName(TableOutputs) + " outputs " +
		"LEFT JOIN " + d.GetTableName(TableInputs) + " inputs ON (inputs.prev_hash = outputs.hash AND inputs.prev_index = outputs.`index`) " +
		"WHERE outputs.address IN (" + db_util.GetQuestionMarksCombined(len(addresses)) + ") " +
		"AND inputs.hash IS NULL"
	var addressStrings = make([]interface{}, len(addresses))
	for i := range addresses {
		addressStrings[i] = addresses[i].String()
	}
	rows, err := d.Db.Query(query, addressStrings...)
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

func (d *Database) SetAddressLastUpdate(lastUpdates []graph.AddressUpdate) error {
	var queries []*Query
	for i := range lastUpdates {
		if lastUpdates[i].Time.Unix() <= 0 {
			continue
		}
		queries = append(queries, d.GetInsert(TableAddressUpdates, map[string]interface{}{
			"address": lastUpdates[i].Address.String(),
			"time":    lastUpdates[i].Time.Unix(),
		}))
	}
	if err := execQueries(d.Db, queries); err != nil {
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
