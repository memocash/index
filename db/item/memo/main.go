package memo

import "github.com/memocash/index/db/item/db"

func GetTopics() []db.Object {
	return []db.Object{
		&PostHeightLike{},
		&LikeTip{},
		&LockHeightFollow{},
		&LockHeightFollowed{},
		&LockHeightLike{},
		&LockHeightName{},
		&LockHeightPost{},
		&LockHeightProfile{},
		&LockHeightProfilePic{},
		&LockHeightRoomFollow{},
		&Post{},
		&PostChild{},
		&PostParent{},
		&PostRoom{},
		&RoomHeightFollow{},
		&RoomHeightPost{},
	}
}
