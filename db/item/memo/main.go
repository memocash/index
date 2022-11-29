package memo

import "github.com/memocash/index/db/item/db"

func GetTopics() []db.Object {
	return []db.Object{
		&PostHeightLike{},
		&LikeTip{},
		&AddrHeightFollow{},
		&AddrHeightFollowed{},
		&AddrHeightLike{},
		&AddrHeightName{},
		&AddrHeightPost{},
		&AddrHeightProfile{},
		&AddrHeightProfilePic{},
		&AddrHeightRoomFollow{},
		&Post{},
		&PostChild{},
		&PostParent{},
		&PostRoom{},
		&RoomHeightFollow{},
		&RoomHeightPost{},
	}
}
