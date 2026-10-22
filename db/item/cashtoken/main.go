// Package cashtoken holds the CashTokens topics: what the index knows about
// token-carrying outputs, decoded from their lock script prefix at save time.
package cashtoken

import "github.com/memocash/index/db/item/db"

func GetTopics() []db.Object {
	return []db.Object{
		&Output{},
	}
}
