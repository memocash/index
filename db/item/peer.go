package item

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type Peer struct {
	Ip       []byte
	Port     uint16
	Services uint64
}

func (p *Peer) GetUid() []byte {
	return jutil.CombineBytes(jutil.GetUintData(uint(p.Port)), p.Ip)
}

func (p *Peer) GetShardSource() uint {
	return client.GenShardSource(p.Ip)
}

func (p *Peer) GetTopic() string {
	return db.TopicPeer
}

func (p *Peer) Serialize() []byte {
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

func GetPeers(ctx context.Context, shard uint32, startId []byte) ([]*Peer, error) {
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	var startIdBytes []byte
	if len(startId) > 0 {
		startIdBytes = startId
	}
	if err := dbClient.GetByPrefix(ctx, db.TopicPeer, client.Prefix{
		Start: startIdBytes,
		Limit: client.LargeLimit,
	}); err != nil {
		return nil, fmt.Errorf("error getting peers from queue client; %w", err)
	}
	var peers = make([]*Peer, len(dbClient.Messages))
	for i := range dbClient.Messages {
		peers[i] = new(Peer)
		db.Set(peers[i], dbClient.Messages[i])
	}
	return peers, nil
}

func GetNextPeer(ctx context.Context, shard uint32, startId []byte) (*Peer, error) {
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetNext(ctx, db.TopicPeer, startId); err != nil {
		return nil, fmt.Errorf("error getting peers from queue client; %w", err)
	} else if len(dbClient.Messages) == 0 {
		return nil, fmt.Errorf("error next peer not found; %w", client.EntryNotFoundError)
	} else if len(dbClient.Messages) != 1 {
		return nil, fmt.Errorf("error unexpected next peer message len (%d)", len(dbClient.Messages))
	}
	var peer = new(Peer)
	db.Set(peer, dbClient.Messages[0])
	return peer, nil
}

func GetCountPeers() (uint64, error) {
	var totalCount uint64
	for _, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		count, err := dbClient.GetTopicCount(db.TopicPeer, nil)
		if err != nil {
			return 0, fmt.Errorf("error getting peer topic count for shard: %d; %w", shardConfig.Shard, err)
		}
		totalCount += count
	}
	return totalCount, nil
}
