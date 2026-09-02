package sql_test

import (
	"database/sql"
	sqldriver "database/sql/driver"
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/mattn/go-sqlite3"
	driver "github.com/memocash/index/client/drivers/sql"
	"github.com/memocash/index/client/lib/graph"
)

func newTestDatabase(t *testing.T) (*driver.Database, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("error opening sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	// database/sql pools connections and each :memory: connection is its
	// own database; pin to one so the schema and the inserts share it.
	db.SetMaxOpenConns(1)
	database, err := driver.NewDatabase(db, "test")
	if err != nil {
		t.Fatalf("error creating database: %v", err)
	}
	return database, db
}

// SLP amounts use the full uint64 range. SaveTxs must persist them exactly,
// including values above math.MaxInt64 that a signed 64-bit column would wrap.
func TestSaveTxsSlpAmountRoundTrip(t *testing.T) {
	tests := []uint64{0, 1000, math.MaxInt64, math.MaxInt64 + 1, math.MaxUint64}
	database, db := newTestDatabase(t)
	var txs []graph.Tx
	for i, amount := range tests {
		txs = append(txs, graph.Tx{
			Hash: "tx" + strconv.Itoa(i),
			Outputs: []graph.Output{{
				Index:  0,
				Amount: 546,
				Slp:    &graph.Slp{Hash: "tx" + strconv.Itoa(i), Index: 0, TokenHash: "token", Amount: amount},
			}},
		})
	}
	if err := database.SaveTxs(txs); err != nil {
		t.Fatalf("error saving txs: %v", err)
	}
	for i, expected := range tests {
		var stored string
		err := db.QueryRow("SELECT amount FROM "+database.GetTableName(driver.TableSlpOutputs)+" WHERE hash = ?",
			"tx"+strconv.Itoa(i)).Scan(&stored)
		if err != nil {
			t.Fatalf("error reading amount %d: %v", expected, err)
		}
		got, err := strconv.ParseUint(stored, 10, 64)
		if err != nil {
			t.Errorf("amount %d stored as %q, not a uint64: %v", expected, stored, err)
			continue
		}
		if got != expected {
			t.Errorf("amount %d stored as %q", expected, stored)
		}
	}
}

// A database created by an earlier release has slp_outputs.amount with INT
// affinity, which wrapped amounts above math.MaxInt64 negative and would turn
// the new decimal strings into imprecise REALs. NewDatabase must rebuild the
// column as text, recover the wrapped rows, and then store new large amounts
// exactly. Running it again must be a no-op.
func TestNewDatabaseMigratesLegacySlpAmount(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("error opening sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	const prefix = "test"
	const legacyTable = prefix + "_slp_outputs"
	if _, err := db.Exec("CREATE TABLE " + legacyTable +
		" (hash CHAR, `index` INT, token_hash CHAR, amount INT, UNIQUE(hash, `index`))"); err != nil {
		t.Fatalf("error creating legacy table: %v", err)
	}
	legacy := []struct {
		hash   string
		stored int64
		want   string
	}{
		{"small", 1000, "1000"},
		{"max_int64", math.MaxInt64, "9223372036854775807"},
		{"wrapped", -1, "18446744073709551615"}, // old int64(math.MaxUint64)
	}
	for _, row := range legacy {
		if _, err := db.Exec("INSERT INTO "+legacyTable+" (hash, `index`, token_hash, amount) VALUES (?, ?, ?, ?)",
			row.hash, 0, "token", row.stored); err != nil {
			t.Fatalf("error inserting legacy row: %v", err)
		}
	}
	for pass := 0; pass < 2; pass++ {
		if _, err := driver.NewDatabase(db, prefix); err != nil {
			t.Fatalf("pass %d: error creating database: %v", pass, err)
		}
	}
	var colType string
	rows, err := db.Query("PRAGMA table_info(" + legacyTable + ")")
	if err != nil {
		t.Fatalf("error reading table info: %v", err)
	}
	for rows.Next() {
		var cid, notNull, pk int
		var name, typ string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk); err != nil {
			t.Fatalf("error scanning table info: %v", err)
		}
		if name == "amount" {
			colType = typ
		}
	}
	_ = rows.Close()
	if colType != "CHAR" {
		t.Fatalf("amount column type after migration: %q, expected CHAR", colType)
	}
	database := &driver.Database{Db: db, Prefix: prefix}
	if err := database.SaveTxs([]graph.Tx{{
		Hash: "new",
		Outputs: []graph.Output{{
			Index: 0, Amount: 546,
			Slp: &graph.Slp{Hash: "new", Index: 0, TokenHash: "token", Amount: math.MaxInt64 + 1},
		}},
	}}); err != nil {
		t.Fatalf("error saving tx after migration: %v", err)
	}
	expected := map[string]string{"new": "9223372036854775808"}
	for _, row := range legacy {
		expected[row.hash] = row.want
	}
	for hash, want := range expected {
		var stored, storageClass string
		if err := db.QueryRow("SELECT amount, typeof(amount) FROM "+legacyTable+" WHERE hash = ?", hash).
			Scan(&stored, &storageClass); err != nil {
			t.Fatalf("error reading %s: %v", hash, err)
		}
		if stored != want || storageClass != "text" {
			t.Errorf("%s: stored %q (%s), expected %q (text)", hash, stored, storageClass, want)
		}
	}
}

// faultDriver wraps go-sqlite3 so a SELECT whose text starts with faultQuery
// returns faultRowsBefore rows and then an iteration error, which
// database/sql surfaces through Rows.Err rather than as a Next/Scan failure.
type faultDriver struct {
	inner sqldriver.Driver
}

const (
	faultDriverName = "sqlite3_migration_fault"
	faultQuery      = "SELECT hash, `index`, token_hash, amount FROM "
	faultRowsBefore = 1
)

var errInjected = errors.New("injected iteration failure")

var registerFaultDriver sync.Once

func (d faultDriver) Open(name string) (sqldriver.Conn, error) {
	conn, err := d.inner.Open(name)
	if err != nil {
		return nil, err
	}
	return faultConn{Conn: conn}, nil
}

// faultConn embeds only the driver.Conn interface, so database/sql does not
// see go-sqlite3's context/queryer fast paths and routes every statement
// through Prepare, where the query text can be inspected.
type faultConn struct {
	sqldriver.Conn
}

func (c faultConn) Prepare(query string) (sqldriver.Stmt, error) {
	stmt, err := c.Conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return faultStmt{Stmt: stmt, query: query}, nil
}

type faultStmt struct {
	sqldriver.Stmt
	query string
}

func (s faultStmt) Query(args []sqldriver.Value) (sqldriver.Rows, error) {
	rows, err := s.Stmt.Query(args)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(s.query, faultQuery) {
		return &faultRows{Rows: rows}, nil
	}
	return rows, nil
}

type faultRows struct {
	sqldriver.Rows
	seen int
}

func (r *faultRows) Next(dest []sqldriver.Value) error {
	if r.seen >= faultRowsBefore {
		return errInjected
	}
	r.seen++
	return r.Rows.Next(dest)
}

// If reading the legacy rows fails part way, the migration must abort and
// leave the legacy table untouched rather than commit a truncated rebuild.
func TestMigrateLegacySlpAmountAbortsOnIterationError(t *testing.T) {
	registerFaultDriver.Do(func() {
		sql.Register(faultDriverName, faultDriver{inner: &sqlite3.SQLiteDriver{}})
	})
	db, err := sql.Open(faultDriverName, ":memory:")
	if err != nil {
		t.Fatalf("error opening sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	const prefix = "test"
	const legacyTable = prefix + "_slp_outputs"
	if _, err := db.Exec("CREATE TABLE " + legacyTable +
		" (hash CHAR, `index` INT, token_hash CHAR, amount INT, UNIQUE(hash, `index`))"); err != nil {
		t.Fatalf("error creating legacy table: %v", err)
	}
	legacy := []int64{1000, math.MaxInt64, -1}
	for i, amount := range legacy {
		if _, err := db.Exec("INSERT INTO "+legacyTable+" (hash, `index`, token_hash, amount) VALUES (?, ?, ?, ?)",
			"tx"+strconv.Itoa(i), 0, "token", amount); err != nil {
			t.Fatalf("error inserting legacy row: %v", err)
		}
	}
	_, err = driver.NewDatabase(db, prefix)
	if !errors.Is(err, errInjected) {
		t.Fatalf("expected migration to fail with injected error, got: %v", err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + legacyTable).Scan(&count); err != nil {
		t.Fatalf("error counting legacy rows: %v", err)
	}
	if count != len(legacy) {
		t.Errorf("legacy table has %d rows after failed migration, expected %d", count, len(legacy))
	}
	for i, amount := range legacy {
		var stored int64
		var storageClass string
		if err := db.QueryRow("SELECT amount, typeof(amount) FROM "+legacyTable+" WHERE hash = ?",
			"tx"+strconv.Itoa(i)).Scan(&stored, &storageClass); err != nil {
			t.Fatalf("error reading legacy row %d: %v", i, err)
		}
		if stored != amount || storageClass != "integer" {
			t.Errorf("legacy row %d: stored %d (%s), expected %d (integer)", i, stored, storageClass, amount)
		}
	}
}
