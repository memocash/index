package network_server

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/act/balance"
	"github.com/memocash/index/node/obj/get"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"github.com/memocash/index/ref/network/gen/network_pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sort"
)

type Server struct {
	listener net.Listener
	grpc     *grpc.Server
	network_pb.UnimplementedNetworkServer
	Verbose bool
	Port    int
}

func (s *Server) OutputMessage(_ context.Context, stringMessage *network_pb.StringMessage) (*network_pb.ErrorReply, error) {
	jlog.Logf("OutputMessage: %s\n", stringMessage.Message)
	return &network_pb.ErrorReply{}, nil
}

func (s *Server) SaveTxs(_ context.Context, txs *network_pb.Txs) (*network_pb.SaveTxsReply, error) {
	var blockTxs = make(map[string][]*wire.MsgTx)
	for _, tx := range txs.Txs {
		txMsg, err := memo.GetMsgFromRaw(tx.Raw)
		if err != nil {
			return nil, jerr.Get("error parsing transaction message", err)
		}
		var blockHashStr string
		if len(tx.Block) > 0 {
			blockHashStr = hex.EncodeToString(tx.Block)
		}
		blockTxs[blockHashStr] = append(blockTxs[blockHashStr], txMsg)
	}
	blockSaver := saver.NewBlock(false)
	combinedSaver := saver.NewCombined([]dbi.TxSave{
		saver.NewTxRaw(false),
		saver.NewTx(false),
		saver.NewUtxo(false),
		saver.NewLockHeight(false),
		saver.NewDoubleSpend(false),
		saver.NewMemo(false),
	})
	for blockHashStr, msgTxs := range blockTxs {
		var blockHeader *wire.BlockHeader
		if blockHashStr != "" {
			blockHash, err := hex.DecodeString(blockHashStr)
			if err != nil {
				return nil, jerr.Getf(err, "error decoding block hash for tx: %s", msgTxs[0].TxHash())
			}
			block, err := item.GetBlock(blockHash)
			if err != nil {
				return nil, jerr.Get("error getting block", err)
			}
			header, err := memo.GetBlockHeaderFromRaw(block.Raw)
			if err != nil {
				return nil, jerr.Get("error getting block header from raw", err)
			}
			blockHeader = header
		}
		if blockHeader != nil {
			if err := blockSaver.SaveBlock(*blockHeader); err != nil {
				return nil, jerr.Get("error saving block", err)
			}
		}
		if err := combinedSaver.SaveTxs(memo.GetBlockFromTxs(msgTxs, blockHeader)); err != nil {
			err = jerr.Get("error saving transactions", err)
			return &network_pb.SaveTxsReply{
				Error: err.Error(),
			}, err
		}
	}
	return &network_pb.SaveTxsReply{}, nil
}

func (s *Server) GetTx(_ context.Context, req *network_pb.TxRequest) (*network_pb.TxReply, error) {
	getTx := get.NewTx(req.Hash)
	if err := getTx.Get(); err != nil {
		return nil, jerr.Get("error getting transaction", err)
	}
	return &network_pb.TxReply{Tx: &network_pb.Tx{
		Raw:   getTx.Raw,
		Block: getTx.BlockHash,
	}}, nil
}

func (s *Server) GetTxBlock(_ context.Context, req *network_pb.TxBlockRequest) (*network_pb.TxBlockReply, error) {
	getTxBlock := get.NewTxBlock()
	if err := getTxBlock.Get(req.Txs); err != nil {
		return nil, jerr.Get("error getting tx blocks by hashes", err)
	}
	var txs = make([]*network_pb.BlockTx, len(getTxBlock.Txs))
	for i := range getTxBlock.Txs {
		txs[i] = &network_pb.BlockTx{
			Block: getTxBlock.Txs[i].BlockHash,
			Tx:    getTxBlock.Txs[i].TxHash,
		}
	}
	return &network_pb.TxBlockReply{Txs: txs}, nil
}

func (s *Server) ListenTx(ctx context.Context, req *network_pb.TxRequest) (*network_pb.ListenTxReply, error) {
	txProcessed, err := item.WaitForTxProcessed(ctx, req.GetHash())
	if err != nil {
		return nil, jerr.Get("error waiting for tx processed", err)
	}
	return &network_pb.ListenTxReply{
		Timestamp: txProcessed.Timestamp.Unix(),
	}, nil
}

func (s *Server) GetOutputInputs(_ context.Context, req *network_pb.OutputInputsRequest) (*network_pb.OutputInputsResponse, error) {
	var outs = make([]memo.Out, len(req.Outputs))
	for i := range req.Outputs {
		outs[i] = memo.Out{
			TxHash: req.Outputs[i].Tx,
			Index:  req.Outputs[i].Index,
		}
	}
	outputInputs, err := item.GetOutputInputs(outs)
	if err != nil {
		return nil, jerr.Get("error getting output inputs", err)
	}
	var inputs = make([]*network_pb.Input, len(outputInputs))
	for i := range outputInputs {
		inputs[i] = &network_pb.Input{
			Tx:          outputInputs[i].Hash,
			Index:       outputInputs[i].Index,
			PrevTxHash:  outputInputs[i].PrevHash,
			PrevTxIndex: outputInputs[i].PrevIndex,
		}
	}
	return &network_pb.OutputInputsResponse{Inputs: inputs}, nil
}

func (s *Server) GetBlockTxs(_ context.Context, req *network_pb.BlockTxRequest) (*network_pb.BlockTxResponse, error) {
	var blockTxs []*network_pb.BlockTx
	for _, shard := range config.GetQueueShards() {
		blockTxsRaw, err := item.GetBlockTxesRaw(item.BlockTxesRawRequest{
			Shard:       shard.Min,
			BlockHash:   req.Block,
			StartTxHash: req.Start,
			Limit:       client.LargeLimit,
		})
		if err != nil {
			return nil, jerr.Get("error getting block txs raw from queue for server", err)
		}
		for _, blockTxRaw := range blockTxsRaw {
			blockTxs = append(blockTxs, &network_pb.BlockTx{
				Block: blockTxRaw.BlockHash,
				Tx:    blockTxRaw.TxHash,
				Raw:   blockTxRaw.Raw,
			})
		}
	}
	return &network_pb.BlockTxResponse{Txs: blockTxs}, nil
}

func (s *Server) GetMempoolTxs(_ context.Context, req *network_pb.MempoolTxRequest) (*network_pb.MempoolTxResponse, error) {
	mempoolTxsRaw, err := item.GetMempoolTxs(req.Start, 0)
	if err != nil {
		return nil, jerr.Get("error getting mempool txs", err)
	}
	var resp = new(network_pb.MempoolTxResponse)
	resp.Txs = make([]*network_pb.MempoolTx, len(mempoolTxsRaw))
	for i := range mempoolTxsRaw {
		resp.Txs[i] = &network_pb.MempoolTx{
			Tx: mempoolTxsRaw[i].TxHash,
		}
	}
	return resp, nil
}

func (s *Server) GetDoubleSpends(_ context.Context, req *network_pb.DoubleSpendRequest) (*network_pb.DoubleSpendResponse, error) {
	doubleSpendOutputs, err := item.GetDoubleSpendOutputs(&item.DoubleSpendOutput{TxHash: req.Start}, 0)
	if err != nil {
		return nil, jerr.Get("error getting mempool txs", err)
	}
	var resp = new(network_pb.DoubleSpendResponse)
	resp.Txs = make([]*network_pb.DoubleSpend, len(doubleSpendOutputs))
	for i := range doubleSpendOutputs {
		resp.Txs[i] = &network_pb.DoubleSpend{
			Tx: doubleSpendOutputs[i].TxHash,
		}
	}
	return resp, nil
}

func (s *Server) GetBalance(_ context.Context, address *network_pb.Address) (*network_pb.BalanceReply, error) {
	bal := balance.NewBalance()
	err := bal.Get(address.GetAddress())
	if err != nil {
		return nil, jerr.Get("error getting balance for address", err)
	}
	return &network_pb.BalanceReply{
		Address:   bal.Address,
		Balance:   bal.Balance,
		Spendable: bal.Spendable,
		Spends:    int32(bal.Spends),
		Utxos:     int32(bal.UtxoCount),
	}, nil
}

func (s *Server) SaveTxBlock(_ context.Context, txBlock *network_pb.TxBlock) (*network_pb.ErrorReply, error) {
	var msgTxs = make([]*wire.MsgTx, len(txBlock.Txs))
	var err error
	for i := range txBlock.Txs {
		if msgTxs[i], err = memo.GetMsgFromRaw(txBlock.Txs[i].Raw); err != nil {
			return nil, jerr.Get("error getting tx from raw", err)
		}
	}
	if blockHeader, err := memo.GetBlockHeaderFromRaw(txBlock.Block.Header); err != nil {
		return nil, jerr.Get("error parsing block header", err)
	} else if err := saver.NewBlock(true).SaveBlock(*blockHeader); err != nil {
		return nil, jerr.Get("error saving block", err)
	} else if err := saver.NewCombined([]dbi.TxSave{
		saver.NewTxRaw(false),
		saver.NewTx(false),
		saver.NewUtxo(false),
		saver.NewLockHeight(false),
		saver.NewDoubleSpend(false),
		saver.NewMemo(false),
	}).SaveTxs(memo.GetBlockFromTxs(msgTxs, blockHeader)); err != nil {
		return nil, jerr.Get("error saving transactions", err)
	}
	return &network_pb.ErrorReply{}, nil
}

func (s *Server) GetBlockInfos(_ context.Context, req *network_pb.BlockRequest) (*network_pb.BlockInfoReply, error) {
	var heightBlocks []*item.HeightBlock
	for _, shardConfig := range config.GetQueueShards() {
		shardHeightBlocks, err := item.GetHeightBlocks(shardConfig.Min, req.GetHeight(), req.Newest)
		if err != nil {
			return nil, jerr.Get("error getting height block raws", err)
		}
		heightBlocks = append(heightBlocks, shardHeightBlocks...)
	}
	if len(heightBlocks) == 0 {
		return nil, jerr.Newf("error no blocks returned for serv get block infos, height: %d", req.GetHeight())
	}
	var blockHashes = make([][]byte, len(heightBlocks))
	for i := range heightBlocks {
		blockHashes[i] = heightBlocks[i].BlockHash
	}
	/*blocks, err := item.GetBlocks(blockHashes)
	if err != nil {
		return nil, jerr.Get("error getting db blocks", err)
	}*/
	var blockInfos = make([]*network_pb.BlockInfo, len(heightBlocks))
	for i := range heightBlocks {
		/*txCount, err := GetBlockTxCount(heightBlocks[i].BlockHash)
		if err != nil {
			return nil, jerr.Get("error getting block tx count", err)
		}*/
		/*var header []byte
		for _, block := range blocks {
			if bytes.Equal(block.Hash, heightBlocks[i].BlockHash) {
				header = block.Raw
				break
			}
		}*/
		blockInfos[i] = &network_pb.BlockInfo{
			Hash:   heightBlocks[i].BlockHash,
			Height: heightBlocks[i].Height,
			//Txs:    txCount,
			//Header: header,
		}
	}
	sort.Slice(blockInfos, func(i, j int) bool {
		if req.Newest {
			return blockInfos[i].Height > blockInfos[j].Height
		}
		return blockInfos[i].Height < blockInfos[j].Height
	})
	return &network_pb.BlockInfoReply{
		Blocks: blockInfos,
	}, nil
}

func (s *Server) GetHeightBlocks(_ context.Context, req *network_pb.BlockHeightRequest) (*network_pb.BlockHeightResponse, error) {
	var response = new(network_pb.BlockHeightResponse)
	if heightBlocks, err := item.GetHeightBlocksAll(req.Start, req.Wait); err != nil {
		return nil, jerr.Get("error getting height blocks all", err)
	} else {
		response.Blocks = make([]*network_pb.BlockHeight, len(heightBlocks))
		for i := range heightBlocks {
			response.Blocks[i] = &network_pb.BlockHeight{
				Height: heightBlocks[i].Height,
				Hash:   heightBlocks[i].BlockHash,
			}
		}
	}
	return response, nil
}

func GetBlockTxCount(blockHash []byte) (int64, error) {
	txCount, err := item.GetBlockTxCount(blockHash)
	if err != nil {
		return 0, jerr.Get("error getting block txes", err)
	}
	return int64(txCount), nil
}

func (s *Server) GetBlockByHash(_ context.Context, req *network_pb.BlockHashRequest) (*network_pb.BlockInfo, error) {
	blockHash := req.GetHash()
	blockHeight, err := item.GetBlockHeight(blockHash)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return nil, jerr.Get("error getting block height by hash", err)
	}
	var height int64
	if blockHeight != nil {
		height = blockHeight.Height
	}
	txCount, err := GetBlockTxCount(blockHash)
	if err != nil {
		return nil, jerr.Get("error getting block tx count", err)
	}
	block, err := item.GetBlock(blockHash)
	if err != nil {
		return nil, jerr.Get("error getting block", err)
	}
	return &network_pb.BlockInfo{
		Hash:   blockHash,
		Height: height,
		Txs:    txCount,
		Header: block.Raw,
	}, nil
}

func (s *Server) GetBlockByHeight(_ context.Context, req *network_pb.BlockRequest) (*network_pb.BlockInfo, error) {
	heightBlock, err := item.GetHeightBlockSingle(req.GetHeight())
	if err != nil {
		return nil, jerr.Get("error getting height block by height", err)
	}
	block, err := item.GetBlock(heightBlock.BlockHash)
	if err != nil {
		return nil, jerr.Get("error getting block", err)
	}
	return &network_pb.BlockInfo{
		Hash:   heightBlock.BlockHash,
		Height: heightBlock.Height,
		Header: block.Raw,
	}, nil
}

func (s *Server) GetUtxos(_ context.Context, req *network_pb.UtxosRequest) (*network_pb.UtxosResponse, error) {
	var utxos []*network_pb.Output
	for _, pkHash := range req.GetPkHashes() {
		lockHash := script.GetLockHashForAddress(wallet.GetAddressFromPkHash(pkHash))
		var lastUid []byte
		for {
			dbUtxos, err := item.GetLockUtxos(lockHash, lastUid)
			if err != nil {
				return nil, jerr.Get("error getting lock outputs", err)
			}
			for _, dbUtxo := range dbUtxos {
				if bytes.Equal(dbUtxo.GetUid(), lastUid) {
					continue
				}
				utxos = append(utxos, &network_pb.Output{
					Tx:     dbUtxo.Hash,
					Index:  dbUtxo.Index,
					Value:  dbUtxo.Value,
					PkHash: pkHash,
				})
			}
			if len(dbUtxos) < client.DefaultLimit {
				break
			}
			lastUid = dbUtxos[len(dbUtxos)-1].GetUid()
		}
	}
	return &network_pb.UtxosResponse{
		Outputs: utxos,
	}, nil
}

func (s *Server) Run() error {
	if err := s.Start(); err != nil {
		return jerr.Get("error starting network server", err)
	}
	// Serve always returns an error
	return jerr.Get("error serving network server", s.Serve())
}

func (s *Server) Start() error {
	var err error
	if s.listener, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.Port)); err != nil {
		return jerr.Get("failed to listen", err)
	}
	s.grpc = grpc.NewServer()
	network_pb.RegisterNetworkServer(s.grpc, s)
	reflection.Register(s.grpc)
	return nil
}

func (s *Server) Serve() error {
	if err := s.grpc.Serve(s.listener); err != nil {
		return jerr.Get("failed to serve", err)
	}
	return jerr.New("network rpc server disconnected")
}

func NewServer(verbose bool, port int) *Server {
	return &Server{
		Verbose: verbose,
		Port:    port,
	}
}
