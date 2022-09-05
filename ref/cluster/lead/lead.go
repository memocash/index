package lead

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"github.com/memocash/index/ref/config"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type Lead struct {
	Port       int
	ShardError chan ShardError
	Error      chan error
	Mutex      sync.Mutex
	Clients    map[int]*Client
	Counter    Counter
}

func (l *Lead) Run() error {
	l.Error = make(chan error)
	l.ShardError = make(chan ShardError)
	l.Clients = make(map[int]*Client)
	clusterShards := config.GetClusterShards()
	for _, clusterShard := range clusterShards {
		conn, err := grpc.Dial(clusterShard.GetHost(), grpc.WithInsecure())
		if err != nil {
			return jerr.Get("error did not connect cluster client", err)
		}
		l.Clients[clusterShard.Int()] = &Client{
			Config: clusterShard,
			Client: cluster_pb.NewClusterClient(conn)}
		go l.StartClient(clusterShard)
	}
	go func() {
		for {
			select {
			case shardError := <-l.ShardError:
				if jerr.HasErrorPart(shardError.Error, "connection refused") {
					jlog.Logf("Shard %d disconnected, waiting for reconnect...\n", shardError.Shard)
					l.Counter.Stop()
					l.Clients[shardError.Shard].Connected = false
					for _, client := range l.Clients {
						if client.Config.Int() == shardError.Shard {
							go l.StartClient(client.Config)
						}
					}
					continue
				}
				l.Error <- jerr.Get("error unhandled from shard", shardError.Error)
			}
		}
	}()
	return jerr.Get("error running lead", <-l.Error)
}

func (l *Lead) CheckAllConnected() {
	for _, client := range l.Clients {
		if !client.Connected {
			return
		}
	}
	jlog.Logf("All shards connected!\n")
	l.Counter.Start(l.Clients, l.ShardError)
}

func (l *Lead) StartClient(cfg config.Shard) {
	for i := 0; ; i++ {
		var client = l.Clients[cfg.Int()]
		resp, err := client.Client.Ping(context.Background(), &cluster_pb.PingReq{
			Nonce: uint64(time.Now().UnixNano()),
		})
		if jerr.HasErrorPart(err, "connection refused") {
			goto Continue
		} else if err != nil {
			l.Error <- jerr.Get("error ping cluster shard", err)
			return
		}
		client.Connected = true
		jlog.Logf("Connected to shard %d, nonce: %d\n", cfg.Int(), resp.Nonce)
		l.CheckAllConnected()
		break
	Continue:
		if i%40 == 0 {
			// Every 10 seconds, depending on timeouts
			jlog.Logf("Waiting for shard %d to start...\n", cfg.Int())
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func NewLead() *Lead {
	return &Lead{}
}
