package memo

import "github.com/memocash/index/db/item/db"

func GetTopics() []db.Object {
	return []db.Object{
		&Liked{},
		&LikeTip{},
		&Post{},
		&PostChild{},
		&PostParent{},
	}
}
