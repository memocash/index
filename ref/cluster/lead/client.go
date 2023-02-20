package lead

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"github.com/memocash/index/ref/config"
	"time"
)

type Client struct {
	Config    config.Shard
	Client    cluster_pb.ClusterClient
	Connected bool
}

func ExecWithRetry(f func() error) error {
	for i := 0; ; i++ {
		if err := f(); jerr.HasErrorPart(err, "connection refused") {
			if i == 0 { // Only first time
				jlog.Logf("Waiting for shard to start...\n")
			}
			time.Sleep(250 * time.Millisecond)
			continue
		} else if err != nil {
			return jerr.Getf(err, "error shard exec with retry")
		}
		return nil
	}
}
