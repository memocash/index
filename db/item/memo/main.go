package memo

import "github.com/memocash/index/db/item/db"

func GetTopics() []db.Object {
	return []db.Object{
		&PostLike{},
		&LikeTip{},
		&AddrFollow{},
		&AddrFollowed{},
		&AddrLike{},
		&AddrName{},
		&AddrPost{},
		&AddrProfile{},
		&AddrProfilePic{},
		&AddrRoomFollow{},
		&Post{},
		&PostChild{},
		&PostParent{},
		&PostRoom{},
		&RoomFollow{},
		&RoomPost{},
		&LinkRequest{},
		&AddrLinkRequest{},
		&AddrLinkRequested{},
	}
}
