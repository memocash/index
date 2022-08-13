package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
)

func (r *roomResolver) Posts(ctx context.Context, obj *model.Room, start *int) ([]*model.Post, error) {
	roomHeightPosts, err := memo.GetRoomHeightPosts(ctx, obj.Name)
	if err != nil {
		return nil, jerr.Get("error getting room height posts for room resolver", err)
	}
	var txHashes = make([][]byte, len(roomHeightPosts))
	for i := range roomHeightPosts {
		txHashes[i] = roomHeightPosts[i].TxHash
	}
	memoPosts, err := memo.GetPosts(txHashes)
	if err != nil {
		return nil, jerr.Get("error getting posts for room resolver", err)
	}
	var posts = make([]*model.Post, len(memoPosts))
	for i := range memoPosts {
		posts[i] = &model.Post{
			TxHash:   hs.GetTxString(memoPosts[i].TxHash),
			LockHash: hex.EncodeToString(memoPosts[i].LockHash),
			Text:     memoPosts[i].Post,
		}
	}
	return posts, nil
}

func (r *roomResolver) Followers(ctx context.Context, obj *model.Room, start *int) ([]*model.Profile, error) {
	panic(fmt.Errorf("not implemented"))
}

// Room returns generated.RoomResolver implementation.
func (r *Resolver) Room() generated.RoomResolver { return &roomResolver{r} }

type roomResolver struct{ *Resolver }
