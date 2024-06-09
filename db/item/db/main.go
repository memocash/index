package db

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
	"sort"
	"sync"
	"time"
)

const (
	TopicFoundPeer      = "found_peer"
	TopicMessage        = "message"
	TopicPeer           = "peer"
	TopicPeerConnection = "peer_connection"
	TopicPeerFound      = "peer_found"
	TopicProcessError   = "process_error"
	TopicProcessStatus  = "process_status"
	TopicSyncStatus     = "sync_status"

	TopicMemoAddrFollow     = "memo_addr_follow"
	TopicMemoAddrFollowed   = "memo_addr_followed"
	TopicMemoAddrLike       = "memo_addr_like"
	TopicMemoAddrName       = "memo_addr_name"
	TopicMemoAddrPost       = "memo_addr_post"
	TopicMemoAddrProfile    = "memo_addr_profile"
	TopicMemoAddrProfilePic = "memo_addr_profile_pic"
	TopicMemoAddrRoomFollow = "memo_addr_room_follow"
	TopicMemoLikeTip        = "memo_like_tip"
	TopicMemoPost           = "memo_post"
	TopicMemoPostChild      = "memo_post_child"
	TopicMemoPostLike       = "memo_post_like"
	TopicMemoPostParent     = "memo_post_parent"
	TopicMemoPostRoom       = "memo_post_room"
	TopicMemoRoomFollow     = "memo_room_follow"
	TopicMemoRoomPost       = "memo_room_post"
	TopicMemoSeenPost       = "memo_seen_post"

	TopicChainBlock           = "chain_block"
	TopicChainBlockHeight     = "chain_block_height"
	TopicChainHeightBlock     = "chain_height_block"
	TopicChainHeightDuplicate = "chain_height_duplicate"
	TopicChainBlockInfo       = "chain_block_info"
	TopicChainBlockTx         = "chain_block_tx"
	TopicChainOutputInput     = "chain_output_input"
	TopicChainTx              = "chain_tx"
	TopicChainTxBlock         = "chain_tx_block"
	TopicChainTxInput         = "chain_tx_input"
	TopicChainTxOutput        = "chain_tx_output"
	TopicChainTxProcessed     = "chain_tx_processed"
	TopicChainTxSeen          = "chain_tx_seen"

	TopicSlpBaton   = "slp_baton"
	TopicSlpGenesis = "slp_genesis"
	TopicSlpMint    = "slp_mint"
	TopicSlpOutput  = "slp_output"
	TopicSlpSend    = "slp_send"

	TopicAddrSeenTx = "addr_seen_tx"
)

type Object interface {
	GetUid() []byte
	GetTopic() string
	GetShardSource() uint
	SetUid(uid []byte)
	Serialize() []byte
	Deserialize(data []byte)
}

func CombineObjects(objectGroups ...[]Object) []Object {
	var objects []Object
	for _, objectGroup := range objectGroups {
		objects = append(objects, objectGroup...)
	}
	return objects
}

func GetShardIdFromByte(b []byte) uint {
	return GetShardId(client.GenShardSource(b))
}

func GetShardIdFromByte32(b []byte) uint32 {
	return uint32(GetShardIdFromByte(b))
}

func GetShardId(shard uint) uint {
	return shard % uint(GetShardCount())
}

func GetShardId32(shard uint) uint32 {
	return uint32(GetShardId(shard))
}

var _shardCount uint32

func GetShardCount() uint32 {
	if _shardCount == 0 {
		configs := config.GetQueueShards()
		if len(configs) > 0 {
			_shardCount = configs[0].Total
		}
	}
	return _shardCount
}

func Save(objects []Object) error {
	var shardMessages = make(map[uint][]*client.Message)
	for i := 0; len(objects) > 0; i++ {
		var object Object
		object, objects = objects[0], objects[1:]
		uid := object.GetUid()
		if len(uid) == 0 {
			uid = make([]byte, 32)
			_, err := rand.Read(uid)
			if err != nil {
				return fmt.Errorf("error getting uid; %w", err)
			}
			object.SetUid(uid)
		}
		shard := GetShardId(object.GetShardSource())
		shardMessages[shard] = append(shardMessages[shard], &client.Message{
			Uid:     uid,
			Message: object.Serialize(),
			Topic:   object.GetTopic(),
		})
	}
	configs := config.GetQueueShards()
	var wg sync.WaitGroup
	wg.Add(len(shardMessages))
	var errs []error
	for shardT, messagesT := range shardMessages {
		go func(shard uint, messages []*client.Message) {
			defer wg.Done()
			sort.Slice(messages, func(i, j int) bool {
				return jutil.ByteLT(messages[i].Uid, messages[j].Uid)
			})
			shardConfig := config.GetShardConfig(uint32(shard), configs)
			queueClient := client.NewClient(shardConfig.GetHost())
			err := queueClient.Save(messages, time.Now())
			if err != nil {
				errs = append(errs, fmt.Errorf("error saving client message; %w", err))
			}
		}(shardT, messagesT)
	}
	wg.Wait()
	if len(errs) > 0 {
		return fmt.Errorf("error saving messages; %w", errors.Join(errs...))
	}
	return nil
}

func Remove(objects []Object) error {
	var shardTopicUids = make(map[uint]map[string][][]byte)
	for _, obj := range objects {
		if shardTopicUids[obj.GetShardSource()] == nil {
			shardTopicUids[obj.GetShardSource()] = make(map[string][][]byte)
		}
		shardTopicUids[obj.GetShardSource()][obj.GetTopic()] =
			append(shardTopicUids[obj.GetShardSource()][obj.GetTopic()], obj.GetUid())
	}
	queueShards := config.GetQueueShards()
	for shard, topicObjects := range shardTopicUids {
		shardConfig := config.GetShardConfig(GetShardId32(shard), queueShards)
		db := client.NewClient(shardConfig.GetHost())
		for topic, uids := range topicObjects {
			if err := db.DeleteMessages(topic, uids); err != nil {
				return fmt.Errorf("error deleting shard topic items: %d %s; %w", shard, topic, err)
			}
		}
	}
	return nil
}

func GetTxHashIndexUid(txHash []byte, index uint32) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(txHash), jutil.GetUint32DataBig(index))
}

func Set(obj Object, msg client.Message) {
	obj.SetUid(msg.Uid)
	obj.Deserialize(msg.Message)
}

type Wait struct {
	Group sync.WaitGroup
	Lock  sync.RWMutex
	Errs  []error
}

func (w *Wait) AddError(err error) {
	w.Lock.Lock()
	defer w.Lock.Unlock()
	w.Errs = append(w.Errs, err)
}

func NewWait(size int) *Wait {
	var wait = new(Wait)
	wait.Group.Add(size)
	return wait
}

type CancelContext struct {
	Context context.Context
	Cancel  func()
}

func NewCancelContext(ctx context.Context, done func()) *CancelContext {
	var c = new(CancelContext)
	c.Context, c.Cancel = context.WithCancel(ctx)
	if done != nil {
		go func() {
			<-c.Context.Done()
			done()
		}()
	}
	return c
}

func FixedTxHashesToRaw(fixedTxHashes [][32]byte) [][]byte {
	var txHashes = make([][]byte, len(fixedTxHashes))
	for i := range fixedTxHashes {
		txHashes[i] = fixedTxHashes[i][:]
	}
	return txHashes
}

func RawTxHashesToFixed(txHashes [][]byte) [][32]byte {
	var fixedTxHashes = make([][32]byte, len(txHashes))
	for i, txHash := range txHashes {
		copy(fixedTxHashes[i][:], txHash)
	}
	return fixedTxHashes
}

func RawTxHashToFixed(txHash []byte) [32]byte {
	var fixedTxHash [32]byte
	copy(fixedTxHash[:], txHash)
	return fixedTxHash
}
