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

func (r *PostRoom) GetShard() uint {
	return client.GetByteShard(r.TxHash[:])
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

func GetPostRoom(ctx context.Context, txHash []byte) (*PostRoom, error) {
	shard := client.GetByteShard32(txHash)
	dbClient := client.NewClient(config.GetShardConfig(shard, config.GetQueueShards()).GetHost())
	err := dbClient.GetSingleContext(ctx, db.TopicMemoPostRoom, jutil.ByteReverse(txHash))
	if err != nil && !client.IsMessageNotSetError(err) {
		return nil, fmt.Errorf("error getting client message post room single; %w", err)
	}
	if len(dbClient.Messages) > 1 {
		return nil, fmt.Errorf("error unexpected number of post room client messages: %d", len(dbClient.Messages))
	} else if len(dbClient.Messages) == 0 {
		return nil, nil
	}
	var postRoom = new(PostRoom)
	db.Set(postRoom, dbClient.Messages[0])
	return postRoom, nil
}
