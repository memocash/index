package slp

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
	"strings"
)

type Genesis struct {
	TxHash     [32]byte
	TokenType  uint8
	Decimals   uint8
	BatonIndex uint32
	Quantity   uint64
	Ticker     string
	Name       string
	DocUrl     string
	DocHash    [32]byte
}

func (g *Genesis) GetTopic() string {
	return db.TopicSlpGenesis
}

func (g *Genesis) GetShard() uint {
	return client.GetByteShard(g.TxHash[:])
}

func (g *Genesis) GetUid() []byte {
	return jutil.ByteReverse(g.TxHash[:])
}

func (g *Genesis) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	copy(g.TxHash[:], jutil.ByteReverse(uid))
}

func (g *Genesis) Serialize() []byte {
	g.Ticker = strings.ReplaceAll(g.Ticker, string([]byte{0x00}), string([]byte{0x01}))
	g.Name = strings.ReplaceAll(g.Name, string([]byte{0x00}), string([]byte{0x01}))
	g.DocUrl = strings.ReplaceAll(g.DocUrl, string([]byte{0x00}), string([]byte{0x01}))
	return jutil.CombineBytes(
		[]byte{g.TokenType, g.Decimals},
		jutil.GetUint32Data(g.BatonIndex),
		jutil.GetUint64Data(g.Quantity),
		g.DocHash[:],
		[]byte(strings.Join([]string{g.Ticker, g.Name, g.DocUrl}, string([]byte{0x00}))),
	)
}

func (g *Genesis) Deserialize(data []byte) {
	if len(data) < 2+4+8+memo.TxHashLength+3 {
		return
	}
	g.TokenType = data[0]
	g.Decimals = data[1]
	g.BatonIndex = jutil.GetUint32(data[2:6])
	g.Quantity = jutil.GetUint64(data[6:14])
	copy(g.DocHash[:], data[14:46])
	split := strings.Split(string(data[46:]), string([]byte{0x00}))
	if len(split) == 3 {
		g.Ticker = split[0]
		g.Name = split[1]
		g.DocUrl = split[2]
	}
}

func GetGenesis(txHash [32]byte) (*Genesis, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(txHash[:]), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(db.TopicSlpGenesis, jutil.ByteReverse(txHash[:])); err != nil {
		return nil, jerr.Get("error getting client message slp genesis", err)
	}
	if len(dbClient.Messages) != 1 {
		return nil, jerr.Newf("error unexpected number of messages slp geneses: %d", len(dbClient.Messages))
	}
	var slpGenesis = new(Genesis)
	db.Set(slpGenesis, dbClient.Messages[0])
	return slpGenesis, nil
}
