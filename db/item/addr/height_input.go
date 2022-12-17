package addr

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type HeightInput struct {
	Addr   [25]byte
	Height int32
	TxHash [32]byte
	Index  uint32
}

func (i *HeightInput) GetTopic() string {
	return db.TopicAddrHeightInput
}

func (i *HeightInput) GetShard() uint {
	return client.GetByteShard(i.Addr[:])
}

func (i *HeightInput) GetUid() []byte {
	return GetHeightTxHashIndexUid(i.Addr, i.Height, i.TxHash, i.Index)
}

func (i *HeightInput) SetUid(uid []byte) {
	if len(uid) != 65 {
		return
	}
	copy(i.Addr[:], uid[:25])
	i.Height = jutil.GetInt32Big(uid[25:29])
	copy(i.TxHash[:], jutil.ByteReverse(uid[29:61]))
	i.Index = jutil.GetUint32Big(uid[61:65])
}

func (i *HeightInput) Serialize() []byte {
	return nil
}

func (i *HeightInput) Deserialize([]byte) {}

func GetHeightInputs(addr [25]byte, start []byte) ([]*HeightInput, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(addr[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicAddrHeightInput,
		Start:    start,
		Prefixes: [][]byte{addr[:]},
		Max:      client.ExLargeLimit,
	}); err != nil {
		return nil, jerr.Get("error getting db height inputs by prefix", err)
	}
	var heightInputs = make([]*HeightInput, len(dbClient.Messages))
	for i := range dbClient.Messages {
		heightInputs[i] = new(HeightInput)
		db.Set(heightInputs[i], dbClient.Messages[i])
	}
	return heightInputs, nil
}

func ListenMempoolAddrHeightInputsMultiple(ctx context.Context, addrs [][25]byte) ([]chan *HeightInput, error) {
	var shardAddrGroups = make(map[uint32][][]byte)
	for _, addr := range addrs {
		shard := db.GetShardByte32(addr[:])
		shardAddrGroups[shard] = append(shardAddrGroups[shard], addr[:])
	}
	var chanHeightInputs []chan *HeightInput
	for shard, addrGroup := range shardAddrGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(ctx, db.TopicAddrHeightInput, addrGroup)
		if err != nil {
			return nil, jerr.Get("error getting addr height input listen message chan", err)
		}
		var chanAddrHeightInput = make(chan *HeightInput)
		go func() {
			for {
				msg, ok := <-chanMessage
				if !ok {
					close(chanAddrHeightInput)
					return
				}
				var addrHeightInput = new(HeightInput)
				db.Set(addrHeightInput, *msg)
				chanAddrHeightInput <- addrHeightInput
			}
		}()
		chanHeightInputs = append(chanHeightInputs, chanAddrHeightInput)
	}
	return chanHeightInputs, nil
}
