package sql

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// migrateSlpAmountText rebuilds slp_outputs when its amount column still has
// the INT affinity from earlier releases. SLP amounts span the full uint64
// range; an INT column wrapped anything above math.MaxInt64 negative (the old
// int64 cast) and turns a large decimal string into an imprecise REAL, so the
// column must have TEXT affinity for the exact decimal written by SaveTxs.
// Idempotent: a no-op once the column type is CHAR.
func migrateSlpAmountText(db *sql.DB, prefix string) error {
	table := tables[TableSlpOutputs]
	name := table.GetName(prefix)
	colType, err := getColumnType(db, name, "amount")
	if err != nil {
		return fmt.Errorf("error getting slp amount column type; %w", err)
	}
	if strings.EqualFold(colType, table.Columns["amount"]) {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting slp amount migration; %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.Query("SELECT hash, `index`, token_hash, amount FROM " + name)
	if err != nil {
		return fmt.Errorf("error reading slp outputs for migration; %w", err)
	}
	var values []map[string]interface{}
	for rows.Next() {
		var hash, tokenHash string
		var index uint32
		var amount interface{}
		if err := rows.Scan(&hash, &index, &tokenHash, &amount); err != nil {
			_ = rows.Close()
			return fmt.Errorf("error scanning slp output for migration; %w", err)
		}
		values = append(values, map[string]interface{}{
			"hash":       hash,
			"index":      index,
			"token_hash": tokenHash,
			"amount":     legacyAmountText(amount),
		})
	}
	// Next returns false on an iteration error as well as at end of data;
	// only Err distinguishes them. Bail before the drop so the rollback
	// keeps the legacy table intact instead of committing a partial copy.
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return fmt.Errorf("error iterating slp outputs for migration; %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("error closing slp outputs for migration; %w", err)
	}
	if _, err := tx.Exec("DROP TABLE " + name); err != nil {
		return fmt.Errorf("error dropping legacy slp outputs table; %w", err)
	}
	if _, err := tx.Exec(table.GetDefinition(prefix)); err != nil {
		return fmt.Errorf("error recreating slp outputs table; %w", err)
	}
	for _, value := range values {
		query := table.GetInsert(prefix, value)
		if _, err := tx.Exec(query.Query, query.Variables...); err != nil {
			return fmt.Errorf("error reinserting slp output for migration; %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing slp amount migration; %w", err)
	}
	return nil
}

// legacyAmountText converts a value read from the INT-affinity column to the
// exact decimal string. A negative integer is an amount the old int64 cast
// wrapped, so the uint64 reinterpretation is the original value. A REAL is
// already lossy (a decimal above math.MaxInt64 stored into the INT column)
// and is kept at its nearest integer.
func legacyAmountText(v interface{}) string {
	switch v := v.(type) {
	case int64:
		return strconv.FormatUint(uint64(v), 10)
	case float64:
		return strconv.FormatFloat(v, 'f', 0, 64)
	case []byte:
		return string(v)
	case string:
		return v
	case nil:
		return "0"
	default:
		return fmt.Sprint(v)
	}
}

func getColumnType(db *sql.DB, table, column string) (string, error) {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return "", fmt.Errorf("error querying table info; %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			return "", fmt.Errorf("error scanning table info; %w", err)
		}
		if name == column {
			return colType, nil
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating table info; %w", err)
	}
	return "", fmt.Errorf("column %s not found in %s", column, table)
}
