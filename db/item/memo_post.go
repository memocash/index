package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type MemoPost struct {
	TxHash   []byte
	LockHash []byte
	Post     string
}

func (p MemoPost) GetUid() []byte {
	return jutil.ByteReverse(p.TxHash)
}

func (p MemoPost) GetShard() uint {
	return client.GetByteShard(p.TxHash)
}

func (p MemoPost) GetTopic() string {
	return TopicMemoPost
}

func (p MemoPost) Serialize() []byte {
	return jutil.CombineBytes(
		p.LockHash,
		[]byte(p.Post),
	)
}

func (p *MemoPost) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	p.TxHash = jutil.ByteReverse(uid)
}

func (p *MemoPost) Deserialize(data []byte) {
	if len(data) < memo.LockHashLength {
		return
	}
	p.LockHash = data[:32]
	p.Post = string(data[32:])
}

func GetMemoPost(txHash []byte) (*MemoPost, error) {
	memoPosts, err := GetMemoPosts([][]byte{txHash})
	if err != nil {
		return nil, jerr.Get("error getting memo posts for single", err)
	}
	if len(memoPosts) == 0 {
		return nil, nil
	}
	return memoPosts[0], nil
}

func GetMemoPosts(txHashes [][]byte) ([]*MemoPost, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := GetShardByte32(txHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(txHash))
	}
	var memoPosts []*MemoPost
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		if err := db.GetByPrefixes(TopicMemoPost, prefixes); err != nil {
			return nil, jerr.Get("error getting client message memo posts", err)
		}
		for _, msg := range db.Messages {
			var memoPost = new(MemoPost)
			Set(memoPost, msg)
			memoPosts = append(memoPosts, memoPost)
		}
	}
	return memoPosts, nil
}
