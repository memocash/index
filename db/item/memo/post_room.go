package memo

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type PostRoom struct {
	TxHash [32]byte
	Room   string
}

func (r *PostRoom) GetTopic() string {
	return db.TopicMemoPostRoom
}

func (r *PostRoom) GetShardSource() uint {
	return client.GenShardSource(r.TxHash[:])
}

func (r *PostRoom) GetUid() []byte {
	return jutil.ByteReverse(r.TxHash[:])
}

func (r *PostRoom) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	copy(r.TxHash[:], jutil.ByteReverse(uid))
}

func (r *PostRoom) Serialize() []byte {
	return []byte(r.Room)
}

func (r *PostRoom) Deserialize(data []byte) {
	r.Room = string(data)
}

func GetPostRooms(ctx context.Context, postTxHashes [][32]byte) ([]*PostRoom, error) {
	messages, err := db.GetSpecific(ctx, db.TopicMemoPostRoom, db.ShardUidsTxHashes(postTxHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting client message memo post rooms; %w", err)
	}
	var postRooms = make([]*PostRoom, len(messages))
	for i := range messages {
		postRooms[i] = new(PostRoom)
		db.Set(postRooms[i], messages[i])
	}
	return postRooms, nil
}
