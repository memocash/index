package shard

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/server"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"strings"
	"time"
)

type ProcessBlock struct {
	Block *wire.MsgBlock
	Added time.Time
}

type Shard struct {
	Id       int
	Verbose  bool
	Error    chan error
	listener net.Listener
	grpc     *grpc.Server
	Blocks   map[chainhash.Hash]ProcessBlock
	cluster_pb.UnimplementedClusterServer
}

func NewShard(shardId int, verbose bool) *Shard {
	return &Shard{
		Id:      shardId,
		Verbose: verbose,
		Blocks:  make(map[chainhash.Hash]ProcessBlock),
	}
}

func (s *Shard) Run() error {
	if err := s.Start(); err != nil {
		return jerr.Get("error starting shard", err)
	}
	return s.Serve()
}

func (s *Shard) Start() error {
	s.Error = make(chan error)
	var err error
	clusterConfig := config.GetShardConfig(uint32(s.Id), config.GetClusterShards())
	if s.listener, err = net.Listen("tcp", clusterConfig.GetHost()); err != nil {
		return jerr.Get("failed to listen cluster shard", err)
	}
	s.grpc = grpc.NewServer(grpc.MaxRecvMsgSize(client.MaxMessageSize), grpc.MaxSendMsgSize(client.MaxMessageSize))
	cluster_pb.RegisterClusterServer(s.grpc, s)
	reflection.Register(s.grpc)
	go func() {
		s.Error <- jerr.Get("failed to serve cluster shard", s.grpc.Serve(s.listener))
	}()
	queueShards := config.GetQueueShards()
	if len(queueShards) < s.Id {
		return jerr.Newf("fatal error shard specified greater than num queue shards: %d %d", s.Id, len(queueShards))
	}
	queueServer := server.NewServer(uint(queueShards[s.Id].Port), uint(s.Id))
	go func() {
		jlog.Logf("Starting queue server shard %d on port %d...\n", queueServer.Shard, queueServer.Port)
		s.Error <- jerr.Getf(queueServer.Run(), "error running queue server for shard: %d", s.Id)
	}()
	return nil
}

func (s *Shard) Serve() error {
	return <-s.Error
}

func (s *Shard) Ping(_ context.Context, req *cluster_pb.PingReq) (*cluster_pb.PingResp, error) {
	jlog.Logf("received ping, nonce: %d\n", req.Nonce)
	return &cluster_pb.PingResp{
		Nonce: uint64(time.Now().UnixNano()),
	}, nil
}

func (s *Shard) SaveTxs(_ context.Context, req *cluster_pb.SaveReq) (*cluster_pb.EmptyResp, error) {
	overallStart := time.Now()
	header, err := memo.GetBlockHeaderFromRaw(req.Block.Header)
	if err != nil {
		return nil, jerr.Get("error getting block header for shard save txs", err)
	}
	var block = &dbi.Block{
		Header:       *header,
		Height:       req.Height,
		Seen:         time.Unix(0, req.Seen),
		Transactions: make([]dbi.Tx, len(req.Block.Txs)),
	}
	var txHashes = make([][32]byte, len(req.Block.Txs))
	for i := range req.Block.Txs {
		msgTx, err := memo.GetMsgFromRaw(req.Block.Txs[i].Raw)
		if err != nil {
			return nil, jerr.Get("error getting msg tx for shard save txs", err)
		}
		block.Transactions[i] = *dbi.WireTxToTx(msgTx, req.Block.Txs[i].Index)
		txHashes[i] = block.Transactions[i].Hash
	}
	seensStart := time.Now()
	var seensDuration time.Duration
	if !req.IsInitial {
		txSeens, err := chain.GetTxSeens(txHashes)
		if err != nil {
			return nil, jerr.Get("error getting tx seens for shard save txs", err)
		}
		seensDuration = time.Since(seensStart)
	TransactionsLoop:
		for i := range block.Transactions {
			for _, txSeen := range txSeens {
				if txSeen.TxHash == block.Transactions[i].Hash {
					block.Transactions[i].Seen = txSeen.Timestamp
					block.Transactions[i].Saved = true
					continue TransactionsLoop
				}
			}
			block.Transactions[i].Seen = block.Seen
		}
	}
	combinedSaver := saver.NewCombinedTx(s.Verbose)
	if err := combinedSaver.SaveTxs(block); err != nil {
		return nil, jerr.Get("error saving block txs shard txs", err)
	}
	if s.Verbose {
		overallDuration := time.Since(overallStart)
		combinedSaver.SaveTimes["seens"] = seensDuration
		combinedSaver.SaveTimes["overall"] = overallDuration
		var saveTimes []string
		for name, duration := range combinedSaver.SaveTimes {
			saveTimes = append(saveTimes, fmt.Sprintf("%s: %s", name, duration))
		}
		jlog.Logf("Save times: %s\n", strings.Join(saveTimes, ", "))
	}
	return &cluster_pb.EmptyResp{}, nil
}
