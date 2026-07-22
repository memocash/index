package attach

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"time"
)

type MemoRoom struct {
	base
	Rooms []*model.Room
}

func ToMemoRooms(ctx context.Context, fields []Field, rooms []*model.Room) error {
	if len(rooms) == 0 {
		return nil
	}
	o := MemoRoom{
		base:  base{Ctx: ctx, Fields: fields},
		Rooms: rooms,
	}
	o.Wait.Add(2)
	go o.AttachPosts()
	go o.AttachFollowers()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo rooms; %w", o.Errors[0])
	}
	return nil
}

func (o *MemoRoom) GetRoomNames() []string {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	var roomNames = make([]string, len(o.Rooms))
	for i := range o.Rooms {
		roomNames[i] = o.Rooms[i].Name
	}
	return roomNames
}

func (o *MemoRoom) getRoomIndexMap() map[string][]int {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	m := make(map[string][]int, len(o.Rooms))
	for i := range o.Rooms {
		m[o.Rooms[i].Name] = append(m[o.Rooms[i].Name], i)
	}
	return m
}

func (o *MemoRoom) AttachPosts() {
	defer o.Wait.Done()
	if !o.HasField([]string{"posts"}) {
		return
	}
	postsField := o.Fields.GetField("posts")
	if txArg, hasTx := postsField.Arguments["tx"]; hasTx && txArg != nil {
		if startArg, hasStart := postsField.Arguments["start"]; !hasStart || startArg == nil {
			o.AddError(fmt.Errorf("room posts tx cursor requires start"))
			return
		}
	}
	startDate, _ := model.UnmarshalDate(postsField.Arguments["start"])
	startTxHash, _ := model.UnmarshalHash(postsField.Arguments["tx"])
	limit, err := unmarshalPageLimit(postsField.Arguments, "room posts", memo.MaxPageSize)
	if err != nil {
		o.AddError(err)
		return
	}
	newest := unmarshalBooleanDefault(postsField.Arguments, "newest", true)
	roomIndexMap := o.getRoomIndexMap()
	if startArg, hasStart := postsField.Arguments["start"]; hasStart && startArg != nil && len(roomIndexMap) > 1 {
		o.AddError(fmt.Errorf("room posts cursor cannot be used with multiple rooms"))
		return
	}
	var allPosts []*model.Post
	for _, roomName := range o.GetRoomNames() {
		roomPosts, err := memo.GetRoomPosts(o.Ctx, roomName, time.Time(startDate), startTxHash, limit, newest)
		if err != nil {
			o.AddError(fmt.Errorf("error getting room height posts for room resolver; %w", err))
			return
		}
		var posts = make([]*model.Post, len(roomPosts))
		for i := range roomPosts {
			posts[i] = &model.Post{TxHash: roomPosts[i].TxHash}
			allPosts = append(allPosts, posts[i])
		}
		o.Mutex.Lock()
		for _, i := range roomIndexMap[roomName] {
			o.Rooms[i].Posts = posts
		}
		o.Mutex.Unlock()
	}
	if err := ToMemoPosts(o.Ctx, postsField.Fields, allPosts); err != nil {
		o.AddError(fmt.Errorf("error attaching to posts for memo rooms; %w", err))
		return
	}
}

func (o *MemoRoom) AttachFollowers() {
	defer o.Wait.Done()
	if !o.HasField([]string{"followers"}) {
		return
	}
	followersField := o.Fields.GetField("followers")
	if txArg, hasTx := followersField.Arguments["tx"]; hasTx && txArg != nil {
		if startArg, hasStart := followersField.Arguments["start"]; !hasStart || startArg == nil {
			o.AddError(fmt.Errorf("room followers tx cursor requires start"))
			return
		}
	}
	startDate, _ := model.UnmarshalDate(followersField.Arguments["start"])
	startTxHash, _ := model.UnmarshalHash(followersField.Arguments["tx"])
	limit, err := unmarshalPageLimit(followersField.Arguments, "room followers", memo.MaxPageSize)
	if err != nil {
		o.AddError(err)
		return
	}
	roomIndexMap := o.getRoomIndexMap()
	if startArg, hasStart := followersField.Arguments["start"]; hasStart && startArg != nil && len(roomIndexMap) > 1 {
		o.AddError(fmt.Errorf("room followers cursor cannot be used with multiple rooms"))
		return
	}
	var allRoomFollows []*model.RoomFollow
	for _, roomName := range o.GetRoomNames() {
		dbRoomFollows, err := memo.GetRoomFollows(o.Ctx, roomName, time.Time(startDate), startTxHash, limit)
		if err != nil {
			o.AddError(fmt.Errorf("error getting room follows for room resolver; %w", err))
			return
		}
		var modelRoomFollows = make([]*model.RoomFollow, len(dbRoomFollows))
		for i := range modelRoomFollows {
			modelRoomFollows[i] = &model.RoomFollow{
				Name:     roomName,
				Address:  dbRoomFollows[i].Addr,
				Unfollow: dbRoomFollows[i].Unfollow,
				TxHash:   dbRoomFollows[i].TxHash,
			}
			allRoomFollows = append(allRoomFollows, modelRoomFollows[i])
		}
		o.Mutex.Lock()
		for _, i := range roomIndexMap[roomName] {
			o.Rooms[i].Followers = modelRoomFollows
		}
		o.Mutex.Unlock()
	}
	if err := ToMemoRoomFollows(o.Ctx, followersField.Fields, allRoomFollows); err != nil {
		o.AddError(fmt.Errorf("error attaching to followers for memo rooms; %w", err))
		return
	}
}
