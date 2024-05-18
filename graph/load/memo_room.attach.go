package load

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
)

type MemoRoomAttach struct {
	baseA
	Rooms []*model.Room
}

func AttachToMemoRooms(ctx context.Context, fields []Field, rooms []*model.Room) error {
	if len(rooms) == 0 {
		return nil
	}
	o := MemoRoomAttach{
		baseA: baseA{Ctx: ctx, Fields: fields},
		Rooms: rooms,
	}
	o.Wait.Add(1)
	go o.AttachPosts()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo rooms; %w", o.Errors[0])
	}
	return nil
}

func (o *MemoRoomAttach) GetRoomNames() []string {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	var roomNames = make([]string, len(o.Rooms))
	for i := range o.Rooms {
		roomNames[i] = o.Rooms[i].Name
	}
	return roomNames
}

func (o *MemoRoomAttach) AttachPosts() {
	defer o.Wait.Done()
	if !o.HasField([]string{"posts"}) {
		return
	}
	var allPosts []*model.Post
	for _, roomName := range o.GetRoomNames() {
		roomPosts, err := memo.GetRoomPosts(o.Ctx, roomName)
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
		for i := range o.Rooms {
			if o.Rooms[i].Name != roomName {
				continue
			}
			o.Rooms[i].Posts = posts
			break
		}
		o.Mutex.Unlock()
	}
	/*if err := AttachToPosts(o.Ctx, GetPrefixFields(o.Fields, "posts."), allOutputs); err != nil {
		o.AddError(fmt.Errorf("error attaching to posts for memo rooms; %w", err))
		return
	}*/
}
