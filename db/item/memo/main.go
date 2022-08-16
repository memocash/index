package memo

import "github.com/memocash/index/db/item/db"

func GetTopics() []db.Object {
	return []db.Object{
		&Liked{},
		&LikeTip{},
		&LockFollow{},
		&LockFollowed{},
		&LockLike{},
		&LockName{},
		&LockPost{},
		&LockProfile{},
		&LockProfilePic{},
		&LockRoomFollow{},
		&Post{},
		&PostChild{},
		&PostParent{},
		&PostRoom{},
		&RoomFollow{},
		&RoomHeightPost{},
	}
}
