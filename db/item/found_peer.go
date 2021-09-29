package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/config"
)

type FoundPeer struct {
	Ip        []byte
	Port      uint16
	FoundIp   []byte
	FoundPort uint16
}

func (p FoundPeer) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.BytePadPrefix(p.Ip, IpBytePadSize),
		jutil.GetUintData(uint(p.Port)),
		jutil.BytePadPrefix(p.FoundIp, IpBytePadSize),
		jutil.GetUintData(uint(p.FoundPort)),
	)
}

func (p FoundPeer) GetShard() uint {
	return client.GetByteShard(p.FoundIp)
}

func (p FoundPeer) GetTopic() string {
	return TopicFoundPeer
}

func (p FoundPeer) Serialize() []byte {
	return nil
}

func (p *FoundPeer) SetUid(uid []byte) {
	if len(uid) != 4*2+IpBytePadSize*2 {
		return
	}
	p.Ip = jutil.ByteUnPad(uid[:IpBytePadSize])
	p.Port = uint16(jutil.GetUint(uid[IpBytePadSize : IpBytePadSize+4]))
	p.FoundIp = jutil.ByteUnPad(uid[IpBytePadSize+4 : IpBytePadSize+4+IpBytePadSize])
	p.FoundPort = uint16(jutil.GetUint(uid[IpBytePadSize+4+IpBytePadSize:]))
}

func (p *FoundPeer) Deserialize([]byte) {}

func GetFoundPeers(shard uint32, startId []byte, ip []byte, port uint16) ([]*FoundPeer, error) {
	var prefix []byte
	if len(ip) > 0 {
		prefix = append(prefix, jutil.BytePadPrefix(ip, IpBytePadSize)...)
		if port > 0 {
			prefix = append(prefix, jutil.GetUintData(uint(port))...)
		}
	}
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var startIdBytes []byte
	if len(startId) > 0 {
		startIdBytes = startId
	}
	opts := client.Opts{
		Topic: TopicFoundPeer,
		Start: startIdBytes,
		Max:   client.LargeLimit,
		Prefixes: [][]byte{prefix},
	}
	if err := dbClient.GetWOpts(opts); err != nil {
		return nil, jerr.Get("error getting found peers from queue client", err)
	}
	var foundPeers = make([]*FoundPeer, len(dbClient.Messages))
	for i := range dbClient.Messages {
		foundPeers[i] = new(FoundPeer)
		foundPeers[i].SetUid(dbClient.Messages[i].Uid)
		foundPeers[i].Deserialize(dbClient.Messages[i].Message)
	}
	return foundPeers, nil
}
