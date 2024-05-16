package memo

import (
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

func (i *SeenPost) GetShard() uint {
	return GetSeenPostShardSource(i.Seen)
}

func GetSeenPostShardSource(seen time.Time) uint {
	return client.GetByteShard(jutil.GetTimeByte(jutil.TimeRoundSeconds(seen, SeenPostShardSeconds)))
}

func GetSeenPostShard32(seen time.Time) uint32 {
	return db.GetShard32(GetSeenPostShardSource(seen))
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

func GetSeenPosts(start time.Time, startTxHash [32]byte) ([]*SeenPost, error) {
	const limit = client.DefaultLimit
	shardConfigs := config.GetQueueShards()
	startShard := GetSeenPostShard32(start)
	var startByte []byte
	if !jutil.IsTimeZero(start) {
		startByte = jutil.CombineBytes(jutil.GetTimeByteNanoBig(start), jutil.ByteReverse(startTxHash[:]))
	}
	var allSeenPosts []*SeenPost
	for i := range shardConfigs {
		shardId := (startShard + uint32(i)) % uint32(len(shardConfigs))
		shardConfig := shardConfigs[shardId]
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic: db.TopicMemoSeenPost,
			Start: startByte,
			Max:   limit,
		}); err != nil {
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
			return allSeenPosts[i].Seen.Before(allSeenPosts[j].Seen)
		})
		if len(allSeenPosts) >= limit && IsSeenPostSameShardWindow(start, allSeenPosts[len(allSeenPosts)-1].Seen) {
			break
		}
	}
	if len(allSeenPosts) > limit {
		allSeenPosts = allSeenPosts[:limit]
	}
	return allSeenPosts, nil
}
