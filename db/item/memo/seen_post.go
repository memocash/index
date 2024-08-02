package memo

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
	"sort"
	"time"
)

// SeenPostShardSeconds shards seen posts into 1 hour groups
const SeenPostShardSeconds = 60 * 60

type SeenPost struct {
	Seen       time.Time
	PostTxHash [32]byte
}

func (i *SeenPost) GetTopic() string {
	return db.TopicMemoSeenPost
}

func (i *SeenPost) GetShardSource() uint {
	return GetSeenPostShardSource(i.Seen)
}

func GetSeenPostShardSource(seen time.Time) uint {
	return client.GenShardSource(jutil.GetTimeByte(jutil.TimeRoundSeconds(seen, SeenPostShardSeconds)))
}

func GetSeenPostShard32(seen time.Time) uint32 {
	return db.GetShardId32(GetSeenPostShardSource(seen))
}

func IsSeenPostSameShardWindow(a, b time.Time) bool {
	return GetSeenPostShardSource(a) == GetSeenPostShardSource(b)
}

func (i *SeenPost) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.GetTimeByteNanoBig(i.Seen),
		jutil.ByteReverse(i.PostTxHash[:]),
	)
}

func (i *SeenPost) SetUid(uid []byte) {
	if len(uid) != 40 {
		return
	}
	i.Seen = jutil.GetByteTimeNanoBig(uid[:8])
	copy(i.PostTxHash[:], jutil.ByteReverse(uid[8:40]))
}

func (i *SeenPost) Serialize() []byte {
	return nil
}

func (i *SeenPost) Deserialize([]byte) {}

func GetSeenPosts(ctx context.Context, start time.Time, startTxHash [32]byte, limit uint32) ([]*SeenPost, error) {
	if limit > client.ExLargeLimit {
		limit = client.ExLargeLimit
	}
	var options = []client.Option{client.OptionNewest()}
	if limit > 0 {
		options = append(options, client.NewOptionLimit(int(limit)))
	}
	startShard := GetSeenPostShard32(start)
	var startByte []byte
	if !jutil.IsTimeZero(start) {
		startByte = jutil.CombineBytes(jutil.GetTimeByteNanoBig(start), jutil.ByteReverse(startTxHash[:]))
	}
	var allSeenPosts []*SeenPost
	for i := range config.GetQueueShards() {
		dbClient := db.GetShardClient(startShard + uint32(i))
		if err := dbClient.GetByPrefix(ctx, db.TopicMemoSeenPost, client.NewStart(startByte), options...); err != nil {
			return nil, fmt.Errorf("error getting db seen posts; %w", err)
		}
		if len(dbClient.Messages) == 0 {
			continue
		}
		var seenPosts = make([]*SeenPost, len(dbClient.Messages))
		for i := range dbClient.Messages {
			seenPosts[i] = new(SeenPost)
			db.Set(seenPosts[i], dbClient.Messages[i])
		}
		allSeenPosts = append(allSeenPosts, seenPosts...)
		sort.Slice(allSeenPosts, func(i, j int) bool {
			if allSeenPosts[i].Seen.Equal(allSeenPosts[j].Seen) {
				return jutil.ByteLT(
					jutil.ByteReverse(allSeenPosts[i].PostTxHash[:]),
					jutil.ByteReverse(allSeenPosts[j].PostTxHash[:]))
			}
			return allSeenPosts[i].Seen.After(allSeenPosts[j].Seen)
		})
		if len(allSeenPosts) >= int(limit) && IsSeenPostSameShardWindow(start, allSeenPosts[len(allSeenPosts)-1].Seen) {
			break
		}
	}
	if len(allSeenPosts) > int(limit) {
		allSeenPosts = allSeenPosts[:limit]
	}
	return allSeenPosts, nil
}
