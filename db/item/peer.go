package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/config"
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

func GetPeers(shard uint32, startId []byte) ([]*Peer, error) {
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var startIdBytes []byte
	if len(startId) > 0 {
		startIdBytes = startId
	}
	err := dbClient.GetLarge(TopicPeer, startIdBytes, false, false)
	if err != nil {
		return nil, jerr.Get("error getting peers from queue client", err)
	}
	var peers = make([]*Peer, len(dbClient.Messages))
	for i := range dbClient.Messages {
		peers[i] = new(Peer)
		peers[i].SetUid(dbClient.Messages[i].Uid)
		peers[i].Deserialize(dbClient.Messages[i].Message)
	}
	return peers, nil
}
