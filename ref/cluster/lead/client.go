package lead

import (
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"github.com/memocash/index/ref/config"
)

type Client struct {
	Config    config.Shard
	Client    cluster_pb.ClusterClient
	Connected bool
}
