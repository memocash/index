package sql

import (
	"database/sql"
	"fmt"
	"strings"
)

func initDb(db *sql.DB, prefix string) error {
	for _, table := range tables {
		if _, err := db.Exec(table.GetDefinition(prefix)); err != nil {
			return fmt.Errorf("error creating table; %w", err)
		}
	}
	return nil
}

func execQueries(db *sql.DB, queries []*Query) error {
	for _, query := range queries {
		if query == nil {
			return fmt.Errorf("exec query is nil")
		}
		if _, err := db.Exec(query.Query, query.Variables...); err != nil {
			return fmt.Errorf("error executing query: %s; %w", query.Name, err)
		}
	}
	return nil
}

type Table struct {
	Name    string
	Columns map[string]string
	Indexes []string
}

func (t Table) GetName(prefix string) string {
	if prefix == "" {
		return t.Name
	}
	return prefix + "_" + t.Name
}

func (t Table) GetInsert(prefix string, values map[string]interface{}) Query {
	var cols []string
	var variables []interface{}
	for col, val := range values {
		if col == "index" {
			col = "`index`"
		}
		cols = append(cols, col)
		variables = append(variables, val)
	}
	return Query{
		Name: t.GetName(prefix),
		Query: "INSERT OR IGNORE INTO " + t.GetName(prefix) +
			" (" + strings.Join(cols, ", ") + ") VALUES (" + "?" + strings.Repeat(", ?", len(cols)-1) + ")",
		Variables: variables,
	}
}

func (t Table) GetDefinition(prefix string) string {
	var columns []string
	for colName, colType := range t.Columns {
		if colName == "index" {
			colName = "`index`"
		}
		columns = append(columns, colName+" "+colType)
	}
	for _, index := range t.Indexes {
		columns = append(columns, index)
	}
	return "CREATE TABLE IF NOT EXISTS " + t.GetName(prefix) + " (" + strings.Join(columns, ", ") + ")"
}

const (
	TableAddressUpdates = "address_updates"
	TableTxs            = "txs"
	TableInputs         = "inputs"
	TableOutputs        = "outputs"
	TableBlocks         = "blocks"
	TableBlockTxs       = "block_txs"
	TableSlpBatons      = "slp_batons"
	TableSlpGeneses     = "slp_geneses"
	TableSlpOutputs     = "slp_outputs"
)

var tables = map[string]Table{
	TableAddressUpdates: {
		Name: TableAddressUpdates,
		Columns: map[string]string{
			"address": "CHAR",
			"time":    "INT",
		},
		Indexes: []string{"UNIQUE(address)"},
	},
	TableTxs: {
		Name: TableTxs,
		Columns: map[string]string{
			"hash": "CHAR",
		},
		Indexes: []string{"UNIQUE(hash)"},
	},
	TableInputs: {
		Name: TableInputs,
		Columns: map[string]string{
			"hash":       "CHAR",
			"index":      "INT",
			"prev_hash":  "CHAR",
			"prev_index": "INT",
		},
		Indexes: []string{"UNIQUE(hash, `index`)"},
	},
	TableOutputs: {
		Name: TableOutputs,
		Columns: map[string]string{
			"hash":    "CHAR",
			"index":   "INT",
			"address": "CHAR",
			"value":   "INT",
		},
		Indexes: []string{"UNIQUE(hash, `index`)"},
	},
	TableBlocks: {
		Name: TableBlocks,
		Columns: map[string]string{
			"hash":      "CHAR",
			"timestamp": "CHAR",
			"height":    "INT",
		},
		Indexes: []string{"UNIQUE(hash)"},
	},
	TableBlockTxs: {
		Name: TableBlockTxs,
		Columns: map[string]string{
			"block_hash": "CHAR",
			"tx_hash":    "CHAR",
		},
		Indexes: []string{"UNIQUE(block_hash, tx_hash)"},
	},
	TableSlpGeneses: {
		Name: TableSlpGeneses,
		Columns: map[string]string{
			"hash":     "CHAR",
			"type":     "INT",
			"decimals": "INT",
			"ticker":   "CHAR",
			"name":     "CHAR",
			"doc_url":  "CHAR",
		},
	},
	TableSlpBatons: {
		Name: TableSlpBatons,
		Columns: map[string]string{
			"hash":       "CHAR",
			"index":      "INT",
			"token_hash": "CHAR",
		},
	},
	TableSlpOutputs: {
		Name: TableSlpOutputs,
		Columns: map[string]string{
			"hash":       "CHAR",
			"index":      "INT",
			"token_hash": "CHAR",
			"amount":     "INT",
		},
	},
}
