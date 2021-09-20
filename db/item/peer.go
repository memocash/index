package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
)

type Peer struct {
	Ip       []byte
	Port     uint16
	Services uint64
}

func (p Peer) GetUid() []byte {
	return jutil.CombineBytes(jutil.GetUintData(uint(p.Port)), p.Ip)
}

func (p Peer) GetShard() uint {
	return client.GetByteShard(p.Ip)
}

func (p Peer) GetTopic() string {
	return TopicPeer
}

func (p Peer) Serialize() []byte {
	return jutil.GetUint64Data(p.Services)
}

func (p *Peer) SetUid(uid []byte) {
	if len(uid) < 4 {
		return
	}
	p.Port = uint16(jutil.GetUint(uid[:4]))
	p.Ip = uid[4:]
}

func (p *Peer) Deserialize(data []byte) {
	p.Services = jutil.GetUint64(data)
}
