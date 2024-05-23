package lead

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"github.com/memocash/index/ref/config"
	"log"
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
				log.Println("Waiting for shard to start...")
			}
			time.Sleep(250 * time.Millisecond)
			continue
		} else if err != nil {
			return fmt.Errorf("error shard exec with retry; %w", err)
		}
		return nil
	}
}
