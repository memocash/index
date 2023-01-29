package sql

import (
	"database/sql"
	"fmt"
)

func initDb(db *sql.DB) error {
	for _, definition := range []string{
		TableAddressUpdates,
		TableBlocks,
		TableBlockTxs,
		TableInputs,
		TableOutputs,
		TableTxs,
	} {
		if _, err := db.Exec(definition); err != nil {
			return fmt.Errorf("%w; error creating table", err)
		}
	}
	return nil
}

func execQueries(db *sql.DB, queries []Query) error {
	for _, query := range queries {
		if _, err := db.Exec(query.Query, query.Variables...); err != nil {
			return fmt.Errorf("%w; error executing query: %s", err, query.Name)
		}
	}
	return nil
}

const tIndex = "`index`"

const (
	TableAddressUpdates = `CREATE TABLE IF NOT EXISTS address_updates (
		address CHAR,
		time INT,
		UNIQUE(address)
    )`
	TableTxs = `CREATE TABLE IF NOT EXISTS txs (
		hash CHAR,
		UNIQUE(hash)
	)`
	TableInputs = `CREATE TABLE IF NOT EXISTS inputs (
        hash CHAR,
        ` + tIndex + ` INT,
        prev_hash CHAR,
        prev_index INT,
        UNIQUE(hash, ` + tIndex + `)
    )`
	TableOutputs = `CREATE TABLE IF NOT EXISTS outputs (
        hash CHAR,
        ` + tIndex + ` INT,
        address CHAR,
        value INT,
        UNIQUE(hash, ` + tIndex + `)
	)`
	TableBlocks = `CREATE TABLE IF NOT EXISTS blocks (
        hash CHAR,
        timestamp CHAR,
        height INT,
        UNIQUE(hash)
    )`
	TableBlockTxs = `CREATE TABLE IF NOT EXISTS block_txs (
        block_hash CHAR,
        tx_hash CHAR,
        UNIQUE(block_hash, tx_hash)
    )`
)
