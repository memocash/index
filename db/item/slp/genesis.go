package slp

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"strings"
)

type Genesis struct {
	TxHash     [32]byte
	Addr       [25]byte
	TokenType  uint8
	Decimals   uint8
	BatonIndex uint32
	Quantity   uint64
	DocHash    [32]byte
	Ticker     string
	Name       string
	DocUrl     string
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
		g.Addr[:],
		[]byte{g.TokenType, g.Decimals},
		jutil.GetUint32Data(g.BatonIndex),
		jutil.GetUint64Data(g.Quantity),
		g.DocHash[:],
		[]byte(strings.Join([]string{g.Ticker, g.Name, g.DocUrl}, string([]byte{0x00}))),
	)
}

func (g *Genesis) Deserialize(data []byte) {
	if len(data) < memo.AddressLength+2+4+8+memo.TxHashLength+3 {
		return
	}
	copy(g.Addr[:], data[:25])
	g.TokenType = data[25]
	g.Decimals = data[26]
	g.BatonIndex = jutil.GetUint32(data[27:31])
	g.Quantity = jutil.GetUint64(data[31:39])
	copy(g.DocHash[:], data[39:71])
	split := strings.Split(string(data[71:]), string([]byte{0x00}))
	if len(split) == 3 {
		g.Ticker = split[0]
		g.Name = split[1]
		g.DocUrl = split[2]
	}
}
