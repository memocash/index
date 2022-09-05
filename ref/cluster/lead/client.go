package lead

import "github.com/memocash/index/ref/cluster/proto/cluster_pb"

type Client struct {
	Client    cluster_pb.ClusterClient
	Connected bool
}
