package slp

import "github.com/memocash/index/db/item/db"

func GetTopics() []db.Object {
	return []db.Object{
		&Genesis{},
		&Mint{},
		&Output{},
		&Baton{},
	}
}
