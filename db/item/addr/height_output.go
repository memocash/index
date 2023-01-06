package addr

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type HeightOutput struct {
	Addr   [25]byte
	Height int32
	TxHash [32]byte
	Index  uint32
	Value  int64
}

func (o *HeightOutput) GetTopic() string {
	return db.TopicAddrHeightOutput
}

func (o *HeightOutput) GetShard() uint {
	return client.GetByteShard(o.Addr[:])
}

func (o *HeightOutput) GetUid() []byte {
	return GetHeightTxHashIndexUid(o.Addr, o.Height, o.TxHash, o.Index)
}

func (o *HeightOutput) SetUid(uid []byte) {
	if len(uid) != 65 {
		return
	}
	copy(o.Addr[:], uid[:25])
	o.Height = jutil.GetInt32Big(uid[25:29])
	copy(o.TxHash[:], jutil.ByteReverse(uid[29:61]))
	o.Index = jutil.GetUint32Big(uid[61:65])
}

func (o *HeightOutput) Serialize() []byte {
	return jutil.GetInt64DataBig(o.Value)
}

func (o *HeightOutput) Deserialize(data []byte) {
	if len(data) != 8 {
		return
	}
	o.Value = jutil.GetInt64Big(data)
}

func GetHeightTxHashIndexUid(addr [25]byte, height int32, txHash [32]byte, index uint32) []byte {
	return jutil.CombineBytes(
		addr[:],
		jutil.GetInt32DataBig(height),
		jutil.ByteReverse(txHash[:]),
		jutil.GetUint32DataBig(index),
	)
}

func GetHeightOutputs(addr [25]byte, start []byte) ([]*HeightOutput, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(addr[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicAddrHeightOutput,
		Start:    start,
		Prefixes: [][]byte{addr[:]},
		Max:      client.ExLargeLimit,
	}); err != nil {
		return nil, jerr.Get("error getting db height outputs by prefix", err)
	}
	var heightOutputs = make([]*HeightOutput, len(dbClient.Messages))
	for i := range dbClient.Messages {
		heightOutputs[i] = new(HeightOutput)
		db.Set(heightOutputs[i], dbClient.Messages[i])
	}
	return heightOutputs, nil
}

func ListenMempoolAddrHeightOutputsMultiple(ctx context.Context, addrs [][25]byte) ([]chan *HeightOutput, error) {
	var shardAddrGroups = make(map[uint32][][]byte)
	for i := range addrs {
		shard := db.GetShardByte32(addrs[i][:])
		shardAddrGroups[shard] = append(shardAddrGroups[shard], addrs[i][:])
	}
	var chanHeightOutputs []chan *HeightOutput
	for shard, addrGroup := range shardAddrGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(ctx, db.TopicAddrHeightOutput, addrGroup)
		if err != nil {
			return nil, jerr.Get("error getting addr height output listen message chan", err)
		}
		var chanAddrHeightOutput = make(chan *HeightOutput)
		go func() {
			for {
				msg, ok := <-chanMessage
				if !ok {
					close(chanAddrHeightOutput)
					return
				}
				var addrHeightOutput = new(HeightOutput)
				db.Set(addrHeightOutput, *msg)
				chanAddrHeightOutput <- addrHeightOutput
			}
		}()
		chanHeightOutputs = append(chanHeightOutputs, chanAddrHeightOutput)
	}
	return chanHeightOutputs, nil
}
