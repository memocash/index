package addr

import "github.com/memocash/index/db/item/db"

func GetTopics() []db.Object {
	return []db.Object{
		&HeightInput{},
		&HeightOutput{},
	}
}
