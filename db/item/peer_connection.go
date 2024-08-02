package item

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
	"time"
)

type PeerConnectionStatus int

func (s PeerConnectionStatus) String() string {
	switch s {
	case PeerConnectionStatusFail:
		return "fail"
	case PeerConnectionStatusSuccess:
		return "success"
	default:
		return "unknown"
	}
}

const (
	PeerConnectionStatusFail    PeerConnectionStatus = 0
	PeerConnectionStatusSuccess PeerConnectionStatus = 1
)

type PeerConnection struct {
	Ip     []byte
	Port   uint16
	Time   time.Time
	Status PeerConnectionStatus
}

func (p *PeerConnection) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.BytePadPrefix(p.Ip, IpBytePadSize),
		jutil.GetUintData(uint(p.Port)),
		jutil.GetTimeByteNanoBig(p.Time),
	)
}

func (p *PeerConnection) GetShardSource() uint {
	return client.GenShardSource(p.Ip)
}

func (p *PeerConnection) GetTopic() string {
	return db.TopicPeerConnection
}

func (p *PeerConnection) Serialize() []byte {
	return jutil.GetIntData(int(p.Status))
}

func (p *PeerConnection) SetUid(uid []byte) {
	if len(uid) != IpBytePadSize+12 {
		return
	}
	p.Ip = jutil.ByteUnPad(uid[:IpBytePadSize])
	p.Port = uint16(jutil.GetUint(uid[IpBytePadSize : IpBytePadSize+4]))
	p.Time = jutil.GetByteTimeNanoBig(uid[IpBytePadSize+4:])
}

func (p *PeerConnection) Deserialize(data []byte) {
	p.Status = PeerConnectionStatus(jutil.GetInt(data))
}

type PeerConnectionsRequest struct {
	Shard   uint32
	StartId []byte
	Ip      []byte
	Port    uint32
	Max     uint32
}

func (r PeerConnectionsRequest) GetShard() uint32 {
	if len(r.Ip) > 0 {
		return client.GenShardSource32(r.Ip)
	}
	return r.Shard
}

func GetPeerConnections(ctx context.Context, request PeerConnectionsRequest) ([]*PeerConnection, error) {
	dbClient := db.GetShardClient(request.GetShard())
	var prefix client.Prefix
	if len(request.Ip) > 0 {
		prefix = client.NewPrefix(jutil.CombineBytes(
			jutil.BytePadPrefix(request.Ip, IpBytePadSize),
			jutil.GetUintData(uint(request.Port)),
		))
	}
	prefix.Limit = request.Max
	if err := dbClient.GetByPrefix(ctx, db.TopicPeerConnection, prefix); err != nil {
		return nil, fmt.Errorf("error getting peer connections from queue client; %w", err)
	}
	var peerConnections = make([]*PeerConnection, len(dbClient.Messages))
	for i := range dbClient.Messages {
		peerConnections[i] = new(PeerConnection)
		db.Set(peerConnections[i], dbClient.Messages[i])
	}
	return peerConnections, nil
}
func GetPeerConnectionLast(ctx context.Context, ip []byte, port uint16) (*PeerConnection, error) {
	peerConnections, err := GetPeerConnectionLasts(ctx, []IpPort{{
		Ip:   ip,
		Port: port,
	}})
	if err != nil {
		return nil, fmt.Errorf("error getting peer connection lasts for single; %w", err)
	}
	if len(peerConnections) != 1 {
		return nil, fmt.Errorf("error did not return expected number of results: %d", len(peerConnections))
	}
	return peerConnections[0], nil
}

type IpPort struct {
	Ip   []byte
	Port uint16
}

func GetPeerConnectionLasts(ctx context.Context, ipPorts []IpPort) ([]*PeerConnection, error) {
	if len(ipPorts) == 0 {
		return nil, nil
	}
	var shardIpPorts = make(map[uint32][]IpPort)
	for _, ipPort := range ipPorts {
		shard := db.GetShardIdFromByte32(ipPort.Ip)
		shardIpPorts[shard] = append(shardIpPorts[shard], ipPort)
	}
	var peerConnections []*PeerConnection
	for shard, ipPorts := range shardIpPorts {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		var prefixes = make([]client.Prefix, len(ipPorts))
		for i := range ipPorts {
			prefixes[i] = client.NewPrefix(jutil.CombineBytes(jutil.BytePadPrefix(ipPorts[i].Ip, IpBytePadSize), jutil.GetUintData(uint(ipPorts[i].Port))))
		}
		opt := client.OptionSinglePrefixLimit()
		if err := dbClient.GetByPrefixes(ctx, db.TopicPeerConnection, prefixes, opt); err != nil {
			return nil, fmt.Errorf("error getting peer connection lasts: %d %d; %w", shard, len(ipPorts), err)
		}
		if len(dbClient.Messages) == 0 {
			return nil, fmt.Errorf("error no peer connection last found; %w", client.EntryNotFoundError)
		}
		for _, message := range dbClient.Messages {
			var peerConnection = new(PeerConnection)
			db.Set(peerConnection, message)
			peerConnections = append(peerConnections, peerConnection)
		}
	}
	return peerConnections, nil
}

func GetCountPeerConnections() (uint64, error) {
	var totalCount uint64
	for _, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		count, err := dbClient.GetTopicCount(db.TopicPeerConnection, nil)
		if err != nil {
			return 0, fmt.Errorf("error getting peer connections topic count for shard: %d; %w", shardConfig.Shard, err)
		}
		totalCount += count
	}
	return totalCount, nil
}
