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

type LinkRequest struct {
	TxHash     [32]byte
	ChildAddr  [25]byte
	ParentAddr [25]byte
	Message    string
}

func (r *LinkRequest) GetTopic() string {
	return db.TopicMemoLinkRequest
}

func (r *LinkRequest) GetShardSource() uint {
	return client.GenShardSource(r.TxHash[:])
}

func (r *LinkRequest) GetUid() []byte {
	return jutil.ByteReverse(r.TxHash[:])
}

func (r *LinkRequest) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		panic("invalid uid size for link request")
	}
	copy(r.TxHash[:], jutil.ByteReverse(uid))
}

func (r *LinkRequest) Serialize() []byte {
	return jutil.CombineBytes(
		r.ChildAddr[:],
		r.ParentAddr[:],
		[]byte(r.Message),
	)
}

func (r *LinkRequest) Deserialize(data []byte) {
	if len(data) < memo.AddressLength*2 {
		panic("invalid data size for link request")
	}
	copy(r.ChildAddr[:], data[:25])
	copy(r.ParentAddr[:], data[25:50])
	r.Message = string(data[50:])
}

func GetLinkRequests(ctx context.Context, txHashes [][32]byte) ([]*LinkRequest, error) {
	var shardUids = make(map[uint32][][]byte)
	for i := range txHashes {
		shard := db.GetShardIdFromByte32(txHashes[i][:])
		shardUids[shard] = append(shardUids[shard], jutil.ByteReverse(txHashes[i][:]))
	}
	var linkRequests []*LinkRequest
	for shard, uids := range shardUids {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Context: ctx,
			Topic:   db.TopicMemoLinkRequest,
			Uids:    uids,
		}); err != nil {
			return nil, fmt.Errorf("error getting client message memo link requests; %w", err)
		}
		for _, msg := range dbClient.Messages {
			var linkRequest = new(LinkRequest)
			db.Set(linkRequest, msg)
			linkRequests = append(linkRequests, linkRequest)
		}
	}
	return linkRequests, nil
}
