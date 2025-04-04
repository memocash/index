package item

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
)

type FoundPeer struct {
	Ip        []byte
	Port      uint16
	FoundIp   []byte
	FoundPort uint16
}

func (p *FoundPeer) GetTopic() string {
	return db.TopicFoundPeer
}

func (p *FoundPeer) GetShardSource() uint {
	return client.GenShardSource(p.FoundIp)
}

func (p *FoundPeer) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.BytePadPrefix(p.Ip, IpBytePadSize),
		jutil.GetUintData(uint(p.Port)),
		jutil.BytePadPrefix(p.FoundIp, IpBytePadSize),
		jutil.GetUintData(uint(p.FoundPort)),
	)
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

func (p *FoundPeer) Serialize() []byte {
	return nil
}

func (p *FoundPeer) Deserialize([]byte) {}

func GetFoundPeers(ctx context.Context, shard uint32, startId []byte, ip []byte, port uint16) ([]*FoundPeer, error) {
	var prefix client.Prefix
	if len(ip) > 0 {
		prefix.Prefix = jutil.BytePadPrefix(ip, IpBytePadSize)
		if port > 0 {
			prefix.Prefix = append(prefix.Prefix, jutil.GetUintData(uint(port))...)
		}
	}
	if len(startId) > 0 {
		prefix.Start = startId
	}
	dbClient := db.GetShardClient(shard)
	if err := dbClient.GetByPrefix(ctx, db.TopicFoundPeer, prefix, client.OptionLargeLimit()); err != nil {
		return nil, fmt.Errorf("error getting found peers from queue client; %w", err)
	}
	var foundPeers = make([]*FoundPeer, len(dbClient.Messages))
	for i := range dbClient.Messages {
		foundPeers[i] = new(FoundPeer)
		db.Set(foundPeers[i], dbClient.Messages[i])
	}
	return foundPeers, nil
}
