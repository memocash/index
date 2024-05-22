package common

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	driver "github.com/memocash/index/client/drivers/sql"
	"github.com/memocash/index/client/lib"
	"github.com/memocash/index/client/lib/graph"
	"log"
	"os"
)

func init() {
	log.SetOutput(os.Stdout)
}

func GetClient() (*lib.Client, error) {
	db, err := sql.Open("sqlite3", "./example.db")
	if err != nil {
		return nil, fmt.Errorf("error opening database; %w", err)
	}
	database, err := driver.NewDatabase(db, "example")
	if err != nil {
		return nil, fmt.Errorf("error getting database; %w", err)
	}
	return lib.NewClient(graph.DefaultUrl, database), nil
}
