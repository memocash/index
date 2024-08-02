package item

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

const (
	IpBytePadSize = 18
)

type PeerFound struct {
	Ip         []byte
	Port       uint16
	FinderIp   []byte
	FinderPort uint16
}

func (p *PeerFound) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.BytePadPrefix(p.Ip, IpBytePadSize),
		jutil.GetUintData(uint(p.Port)),
		jutil.BytePadPrefix(p.FinderIp, IpBytePadSize),
		jutil.GetUintData(uint(p.FinderPort)),
	)
}

func (p *PeerFound) GetShardSource() uint {
	return client.GenShardSource(p.Ip)
}

func (p *PeerFound) GetTopic() string {
	return db.TopicPeerFound
}

func (p *PeerFound) Serialize() []byte {
	return nil
}

func (p *PeerFound) SetUid(uid []byte) {
	if len(uid) != 4*2+IpBytePadSize*2 {
		return
	}
	p.Ip = jutil.ByteUnPad(uid[:IpBytePadSize])
	p.Port = uint16(jutil.GetUint(uid[IpBytePadSize : IpBytePadSize+4]))
	p.FinderIp = jutil.ByteUnPad(uid[IpBytePadSize+4 : IpBytePadSize+4+IpBytePadSize])
	p.FinderPort = uint16(jutil.GetUint(uid[IpBytePadSize+4+IpBytePadSize:]))
}

func (p *PeerFound) Deserialize([]byte) {}

func GetPeerFounds(ctx context.Context, shard uint32, startId []byte) ([]*PeerFound, error) {
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var startIdBytes []byte
	if len(startId) > 0 {
		startIdBytes = startId
	}
	if err := dbClient.GetByPrefix(ctx, db.TopicPeerFound, client.Prefix{
		Start: startIdBytes,
		Limit: client.LargeLimit,
	}); err != nil {
		return nil, fmt.Errorf("error getting peer founds from queue client; %w", err)
	}
	var peerFounds = make([]*PeerFound, len(dbClient.Messages))
	for i := range dbClient.Messages {
		peerFounds[i] = new(PeerFound)
		db.Set(peerFounds[i], dbClient.Messages[i])
	}
	return peerFounds, nil
}
