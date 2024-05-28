package memo

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
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
	var shardUids = make(map[uint32][][]byte)
	for i := range postTxHashes {
		shard := db.GetShardIdFromByte32(postTxHashes[i][:])
		shardUids[shard] = append(shardUids[shard], jutil.ByteReverse(postTxHashes[i][:]))
	}
	var postRooms []*PostRoom
	for shard, uids := range shardUids {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Context:  ctx,
			Topic:    db.TopicMemoPostRoom,
			Uids: uids,
		}); err != nil {
			return nil, fmt.Errorf("error getting client message memo post rooms; %w", err)
		}
		for _, msg := range dbClient.Messages {
			var postRoom = new(PostRoom)
			db.Set(postRoom, msg)
			postRooms = append(postRooms, postRoom)
		}
	}
	return postRooms, nil
}
