package db

import (
	"database/sql"
	"github.com/jchavannes/jgo/jerr"
	_ "github.com/mattn/go-sqlite3"
	"github.com/memocash/index/client/lib"
)

func GetClient() (*lib.Client, error) {
	database, err := NewDatabase()
	if err != nil {
		return nil, jerr.Get("error getting database", err)
	}
	return lib.NewClient("http://localhost:26770/graphql", database), nil
}

func initDb(db *sql.DB) error {
	for _, definition := range []string{
		TableBlocks,
		TableBlockTxs,
		TableInputs,
		TableOutputs,
		TableTxs,
	} {
		if _, err := db.Exec(definition); err != nil {
			return jerr.Get("error creating table", err)
		}
	}
	return nil
}

func execQueries(db *sql.DB, queries []Query) error {
	for _, query := range queries {
		if _, err := db.Exec(query.Query, query.Variables...); err != nil {
			return jerr.Getf(err, "error executing query: %s", query.Name)
		}
	}
	return nil
}

const tIndex = "`index`"

const (
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
