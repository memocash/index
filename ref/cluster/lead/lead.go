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
	Processor  *Processor
	Verbose    bool
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
	l.Processor = NewProcessor(l.Verbose, l.Clients, l.ShardError)
	go func() {
		for {
			select {
			case shardError := <-l.ShardError:
				if jerr.HasErrorPart(shardError.Error, "connection refused") || // Dead connection
					jerr.HasErrorPart(shardError.Error, "error reading from server: EOF") { // Died in middle of request
					jlog.Logf("Shard %d disconnected, waiting for reconnect...\n", shardError.Shard)
					l.Processor.Stop()
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

func (l *Lead) CheckAllConnected() error {
	for _, client := range l.Clients {
		if !client.Connected {
			return nil
		}
	}
	jlog.Logf("All shards connected!\n")
	if err := l.Processor.Start(); err != nil {
		return jerr.Get("error starting processor", err)
	}
	return nil
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
		if err := l.CheckAllConnected(); err != nil {
			l.Error <- jerr.Get("error checking all connected", err)
			return
		}
		break
	Continue:
		if i == 0 {
			// Only first time
			jlog.Logf("Waiting for shard %d to start...\n", cfg.Int())
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func NewLead(verbose bool) *Lead {
	return &Lead{
		Verbose: verbose,
	}
}
