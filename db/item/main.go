package item

import (
	"context"
	"crypto/rand"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
	"sort"
	"sync"
	"time"
)

const (
	TopicBlock                 = "block"
	TopicBlockHeight           = "block_height"
	TopicBlockTx               = "block_tx"
	TopicBlockTxRaw            = "block_tx_raw"
	TopicDoubleSpendOutput     = "double_spend_output"
	TopicDoubleSpendInput      = "double_spend_input"
	TopicFoundPeer             = "found_peer"
	TopicHeightBlock           = "height_block"
	TopicHeightBlockShard      = "height_block_shard"
	TopicHeightDuplicate       = "height_duplicate"
	TopicHeightProcessed       = "height_processed"
	TopicLockAddress           = "lock_address"
	TopicLockBalance           = "lock_balance"
	TopicLockOutput            = "lock_output"
	TopicLockHeightOutput      = "lock_height_output"
	TopicLockHeightOutputInput = "lock_height_output_input"
	TopicLockUtxo              = "lock_utxo"
	TopicLockUtxoLost          = "lock_utxo_lost"
	TopicMempoolTxRaw          = "mempool_tx_raw"
	TopicMessage               = "message"
	TopicOutputInput           = "output_input"
	TopicPeer                  = "peer"
	TopicPeerConnection        = "peer_connection"
	TopicPeerFound             = "peer_found"
	TopicProcessStatus         = "process_status"
	TopicTx                    = "tx"
	TopicTxBlock               = "tx_block"
	TopicTxInput               = "tx_input"
	TopicTxLost                = "tx_lost"
	TopicTxOutput              = "tx_output"
	TopicTxProcessed           = "tx_processed"
	TopicTxSeen                = "tx_seen"
	TopicTxSuspect             = "tx_suspect"
	TopicDoubleSpendSeen       = "double_spend_seen"
	TopicMemoName              = "memo_name"
	TopicMemoProfile           = "memo_profile"
	TopicMemoProfilePic        = "memo_profile_pic"
	TopicMemoFollow            = "memo_follow"
	TopicMemoFollowed          = "memo_followed"
	TopicMemoPost              = "memo_post"
)

type Object interface {
	GetUid() []byte
	GetTopic() string
	GetShard() uint
	SetUid(uid []byte)
	Serialize() []byte
	Deserialize(data []byte)
}

func GetShardByte(b []byte) uint {
	return GetShard(client.GetByteShard(b))
}

func GetShardByte32(b []byte) uint32 {
	return uint32(GetShardByte(b))
}

func GetShard(shard uint) uint {
	return shard % uint(GetShardCount())
}

func GetShard32(shard uint) uint32 {
	return uint32(GetShard(shard))
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
				return jerr.Get("error getting uid", err)
			}
			object.SetUid(uid)
		}
		shard := GetShard(object.GetShard())
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
				errs = append(errs, jerr.Get("error saving client message", err))
			}
		}(shardT, messagesT)
	}
	wg.Wait()
	if len(errs) > 0 {
		return jerr.Get("error saving messages", jerr.Combine(errs...))
	}
	return nil
}

func GetTxHashIndexUid(txHash []byte, index uint32) []byte {
	return jutil.CombineBytes(jutil.ByteReverse(txHash), jutil.GetUint32Data(index))
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

func GetTopics() []Object {
	return []Object{
		&Block{},
		&BlockHeight{},
		&BlockTx{},
		&DoubleSpendInput{},
		&DoubleSpendOutput{},
		&DoubleSpendSeen{},
		&FoundPeer{},
		&HeightBlock{},
		&HeightBlockShard{},
		&HeightDuplicate{},
		&HeightProcessed{},
		&LockAddress{},
		&LockBalance{},
		&LockHeightOutput{},
		&LockHeightOutputInput{},
		&LockOutput{},
		&LockUtxo{},
		&LockUtxoLost{},
		&MemoFollow{},
		&MemoFollowed{},
		&MemoName{},
		&MemoPost{},
		&MemoProfile{},
		&MemoProfilePic{},
		&MempoolTxRaw{},
		&Message{},
		&OutputInput{},
		&Peer{},
		&PeerConnection{},
		&PeerFound{},
		&ProcessStatus{},
		&Tx{},
		&TxBlock{},
		&TxInput{},
		&TxLost{},
		&TxOutput{},
		&TxProcessed{},
		&TxSeen{},
		&TxSuspect{},
	}
}
