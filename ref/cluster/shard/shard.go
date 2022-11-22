package shard

import (
	"context"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/server"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/cluster/proto/cluster_pb"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
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

func (s *Shard) CheckProcessBlocks() {
	for blockHash, block := range s.Blocks {
		if time.Since(block.Added) > time.Minute*5 {
			jlog.Logf("block not processed, removing from shard: %s\n", blockHash)
			delete(s.Blocks, blockHash)
		}
	}
	for len(s.Blocks) > 10 {
		var oldest ProcessBlock
		for _, block := range s.Blocks {
			if oldest.Added.IsZero() || block.Added.Before(oldest.Added) {
				oldest = block
			}
		}
		jlog.Logf("too many blocks in shard, removing oldest: %s\n", oldest.Block.BlockHash())
		delete(s.Blocks, oldest.Block.BlockHash())
	}
}

func (s *Shard) Run() error {
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
	return <-s.Error
}

func (s *Shard) Ping(_ context.Context, req *cluster_pb.PingReq) (*cluster_pb.PingResp, error) {
	jlog.Logf("received ping, nonce: %d\n", req.Nonce)
	return &cluster_pb.PingResp{
		Nonce: uint64(time.Now().UnixNano()),
	}, nil
}

func (s *Shard) SaveTxs(_ context.Context, req *cluster_pb.SaveReq) (*cluster_pb.EmptyResp, error) {
	header, err := memo.GetBlockHeaderFromRaw(req.Block.Header)
	if err != nil {
		return nil, jerr.Get("error getting block header", err)
	}
	var block = &dbi.Block{
		Header: *header,
	}
	for _, tx := range req.Block.Txs {
		msgTx, err := memo.GetMsgFromRaw(tx.Raw)
		if err != nil {
			return nil, jerr.Get("error getting msg tx", err)
		}
		block.Transactions = append(block.Transactions, *dbi.WireTxToTx(msgTx, tx.Index))
	}
	txSaver := saver.NewCombined([]dbi.TxSave{
		saver.NewTxMinimal(s.Verbose),
		saver.NewAddress(s.Verbose),
		saver.NewMemo(s.Verbose),
	})
	if err := txSaver.SaveTxs(block); err != nil {
		return nil, jerr.Get("error saving block txs shard txs", err)
	}
	return &cluster_pb.EmptyResp{}, nil
}

func (s *Shard) process(blockHashByte []byte, initialSync bool) error {
	blockHash, err := chainhash.NewHash(blockHashByte)
	if err != nil {
		return jerr.Get("error parsing block hash for shard save utxos", err)
	}
	block, err := item.GetBlock(blockHash[:])
	if err != nil {
		return jerr.Get("error getting block for shard save utxos", err)
	}
	blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
	if err != nil {
		return jerr.Get("error getting block header from raw for shard save utxos", err)
	}
	var startTxHash []byte
	for {
		blockTxsRaw, err := item.GetBlockTxesRaw(item.BlockTxesRawRequest{
			Shard:       uint32(s.Id),
			BlockHash:   blockHash[:],
			StartTxHash: startTxHash,
			Limit:       client.ExLargeLimit,
		})
		if err != nil {
			return jerr.Get("error getting block txs raw for shard processor", err)
		}
		var txs = make([]*wire.MsgTx, len(blockTxsRaw))
		for i := range blockTxsRaw {
			txs[i], err = memo.GetMsgFromRaw(blockTxsRaw[i].Raw)
			if err != nil {
				return jerr.Get("error getting msg tx from raw for process shard utxos", err)
			}
		}
		txSaver := saver.NewCombined([]dbi.TxSave{
			saver.NewTxMinimal(s.Verbose),
			saver.NewMemo(s.Verbose),
		})
		if err := txSaver.SaveTxs(dbi.WireBlockToBlock(memo.GetBlockFromTxs(txs, blockHeader))); err != nil {
			return jerr.Get("error saving block txs shard utxos", err)
		}
		if len(blockTxsRaw) == client.ExLargeLimit {
			startTxHash = blockTxsRaw[len(blockTxsRaw)-1].TxHash
		} else {
			break
		}
	}
	return nil
}

func (s *Shard) ProcessInitial(_ context.Context, req *cluster_pb.ProcessReq) (*cluster_pb.EmptyResp, error) {
	if err := s.process(req.BlockHash, true); err != nil {
		return nil, jerr.Get("error processing block initial for shard", err)
	}
	return &cluster_pb.EmptyResp{}, nil
}

func (s *Shard) Process(_ context.Context, req *cluster_pb.ProcessReq) (*cluster_pb.EmptyResp, error) {
	if err := s.process(req.BlockHash, false); err != nil {
		return nil, jerr.Get("error processing block for shard", err)
	}
	return &cluster_pb.EmptyResp{}, nil
}

func NewShard(shardId int, verbose bool) *Shard {
	return &Shard{
		Id:      shardId,
		Verbose: verbose,
		Blocks:  make(map[chainhash.Hash]ProcessBlock),
	}
}
