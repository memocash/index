package network_server

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/act/tx_raw"
	"github.com/memocash/index/node/obj/get"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"github.com/memocash/index/ref/network/gen/network_pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
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
	log.Printf("OutputMessage: %s\n", stringMessage.Message)
	return &network_pb.ErrorReply{}, nil
}

func (s *Server) SaveTxs(ctx context.Context, txs *network_pb.Txs) (*network_pb.SaveTxsReply, error) {
	var blockTxs = make(map[string][]*wire.MsgTx)
	for _, tx := range txs.Txs {
		txMsg, err := memo.GetMsgFromRaw(tx.Raw)
		if err != nil {
			return nil, fmt.Errorf("error parsing transaction message; %w", err)
		}
		var blockHashStr string
		if len(tx.Block) > 0 {
			blockHashStr = hex.EncodeToString(tx.Block)
		}
		blockTxs[blockHashStr] = append(blockTxs[blockHashStr], txMsg)
	}
	blockSaver := saver.NewBlock(false)
	combinedSaver := saver.NewCombinedTx(false)
	for blockHashStr, msgTxs := range blockTxs {
		var blockHeader *wire.BlockHeader
		if blockHashStr != "" {
			blockHash, err := chainhash.NewHashFromStr(blockHashStr)
			if err != nil {
				return nil, fmt.Errorf("error decoding block hash for tx: %s; %w", msgTxs[0].TxHash(), err)
			}
			block, err := chain.GetBlock(*blockHash)
			if err != nil {
				return nil, fmt.Errorf("error getting block; %w", err)
			}
			header, err := memo.GetBlockHeaderFromRaw(block.Raw)
			if err != nil {
				return nil, fmt.Errorf("error getting block header from raw; %w", err)
			}
			blockHeader = header
		}
		if blockHeader != nil {
			if err := blockSaver.SaveBlock(dbi.BlockInfo{Header: *blockHeader}); err != nil {
				return nil, fmt.Errorf("error saving block; %w", err)
			}
		}
		if err := combinedSaver.SaveTxs(ctx, dbi.WireBlockToBlock(memo.GetBlockFromTxs(msgTxs, blockHeader))); err != nil {
			err = fmt.Errorf("error saving transactions; %w", err)
			return &network_pb.SaveTxsReply{
				Error: err.Error(),
			}, err
		}
	}
	return &network_pb.SaveTxsReply{}, nil
}

func (s *Server) GetTx(ctx context.Context, req *network_pb.TxRequest) (*network_pb.TxReply, error) {
	getTx := get.NewTx(db.RawTxHashToFixed(req.Hash))
	if err := getTx.Get(ctx); err != nil {
		return nil, fmt.Errorf("error getting transaction; %w", err)
	}
	return &network_pb.TxReply{Tx: &network_pb.Tx{
		Raw:   getTx.Raw,
		Block: getTx.BlockHash[:],
	}}, nil
}

func (s *Server) GetTxBlock(ctx context.Context, req *network_pb.TxBlockRequest) (*network_pb.TxBlockReply, error) {
	txHashes := db.RawTxHashesToFixed(req.Txs)
	chainTxs, err := chain.GetTxBlocks(ctx, txHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting tx blocks from queue for server; %w", err)
	}
	var txs = make([]*network_pb.BlockTx, len(chainTxs))
	for i := range chainTxs {
		txs[i] = &network_pb.BlockTx{
			Block: chainTxs[i].BlockHash[:],
			Tx:    chainTxs[i].TxHash[:],
		}
	}
	return &network_pb.TxBlockReply{Txs: txs}, nil
}

func (s *Server) ListenTx(ctx context.Context, req *network_pb.TxRequest) (*network_pb.ListenTxReply, error) {
	txProcessed, err := chain.WaitForTxProcessed(ctx, req.GetHash())
	if err != nil {
		return nil, fmt.Errorf("error waiting for tx processed; %w", err)
	}
	return &network_pb.ListenTxReply{
		Timestamp: txProcessed.Timestamp.Unix(),
	}, nil
}

func (s *Server) GetOutputInputs(ctx context.Context, req *network_pb.OutputInputsRequest) (*network_pb.OutputInputsResponse, error) {
	var outs = make([]memo.Out, len(req.Outputs))
	for i := range req.Outputs {
		outs[i] = memo.Out{
			TxHash: req.Outputs[i].Tx,
			Index:  req.Outputs[i].Index,
		}
	}
	outputInputs, err := chain.GetOutputInputs(ctx, outs)
	if err != nil {
		return nil, fmt.Errorf("error getting output inputs; %w", err)
	}
	var inputs = make([]*network_pb.Input, len(outputInputs))
	for i := range outputInputs {
		inputs[i] = &network_pb.Input{
			Tx:          outputInputs[i].Hash[:],
			Index:       outputInputs[i].Index,
			PrevTxHash:  outputInputs[i].PrevHash[:],
			PrevTxIndex: outputInputs[i].PrevIndex,
		}
	}
	return &network_pb.OutputInputsResponse{Inputs: inputs}, nil
}

func (s *Server) GetBlockTxs(_ context.Context, req *network_pb.BlockTxRequest) (*network_pb.BlockTxResponse, error) {
	var blockTxs []*network_pb.BlockTx
	// TODO: Reimplement if needed
	return &network_pb.BlockTxResponse{Txs: blockTxs}, nil
}

func (s *Server) GetMempoolTxs(_ context.Context, req *network_pb.MempoolTxRequest) (*network_pb.MempoolTxResponse, error) {
	var resp = new(network_pb.MempoolTxResponse)
	// TODO: Reimplement if needed
	return resp, nil
}

func (s *Server) GetDoubleSpends(_ context.Context, req *network_pb.DoubleSpendRequest) (*network_pb.DoubleSpendResponse, error) {
	var resp = new(network_pb.DoubleSpendResponse)
	// TODO: Reimplement if needed
	return resp, nil
}

func (s *Server) GetBalance(_ context.Context, address *network_pb.Address) (*network_pb.BalanceReply, error) {
	// TODO: Reimplement if needed
	return &network_pb.BalanceReply{}, nil
}

func (s *Server) SaveTxBlock(ctx context.Context, txBlock *network_pb.TxBlock) (*network_pb.ErrorReply, error) {
	var msgTxs = make([]*wire.MsgTx, len(txBlock.Txs))
	var err error
	for i := range txBlock.Txs {
		if msgTxs[i], err = memo.GetMsgFromRaw(txBlock.Txs[i].Raw); err != nil {
			return nil, fmt.Errorf("error getting tx from raw; %w", err)
		}
	}
	blockHeader, err := memo.GetBlockHeaderFromRaw(txBlock.Block.Header)
	if err != nil {
		return nil, fmt.Errorf("error parsing block header; %w", err)
	}
	blockSaver := saver.NewBlock(true)
	if err := blockSaver.SaveBlock(dbi.BlockInfo{Header: *blockHeader}); err != nil {
		return nil, fmt.Errorf("error saving block; %w", err)
	}
	blockTxs, err := chain.GetBlockTxs(chain.BlockTxsRequest{BlockHash: blockHeader.BlockHash()})
	if err != nil {
		return nil, fmt.Errorf("error getting existing block txes for saving block; %w", err)
	}
	sort.Slice(blockTxs, func(i, j int) bool {
		return blockTxs[i].Index < blockTxs[j].Index
	})
	var txHashes = make([][32]byte, len(blockTxs))
	for i := range blockTxs {
		txHashes[i] = blockTxs[i].TxHash
	}
	txRaws, err := tx_raw.Get(ctx, txHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting existing tx raws for tx block saver; %w", err)
	}
	var existingMsgTxs []*wire.MsgTx
BlockTxsLoop:
	for _, blockTx := range blockTxs {
		for _, txRaw := range txRaws {
			if txRaw.Hash == blockTx.TxHash {
				txMsg, err := memo.GetMsgFromRaw(txRaw.Raw)
				if err != nil {
					return nil, fmt.Errorf("error getting message from existing tx raw for tx block saver; %w", err)
				}
				existingMsgTxs = append(existingMsgTxs, txMsg)
				continue BlockTxsLoop
			}
		}
		return nil, fmt.Errorf("error missing tx raw for blockTx: %s", chainhash.Hash(blockTx.TxHash))
	}
	msgTxs = append(existingMsgTxs, msgTxs...)

	block := dbi.WireBlockToBlock(memo.GetBlockFromTxs(msgTxs, blockHeader))
	block.Height = blockSaver.NewHeight
	if err := saver.NewCombinedTx(false).SaveTxs(ctx, block); err != nil {
		return nil, fmt.Errorf("error saving transactions; %w", err)
	}
	return &network_pb.ErrorReply{}, nil
}

func (s *Server) GetBlockInfos(_ context.Context, req *network_pb.BlockRequest) (*network_pb.BlockInfoReply, error) {
	var heightBlocks []*chain.HeightBlock
	for _, shardConfig := range config.GetQueueShards() {
		shardHeightBlocks, err := chain.GetHeightBlocks(shardConfig.Shard, req.GetHeight(), req.Newest)
		if err != nil {
			return nil, fmt.Errorf("error getting height block raws; %w", err)
		}
		heightBlocks = append(heightBlocks, shardHeightBlocks...)
	}
	if len(heightBlocks) == 0 {
		return nil, fmt.Errorf("error no blocks returned for serv get block infos, height: %d", req.GetHeight())
	}
	var blockHashes = make([][]byte, len(heightBlocks))
	for i := range heightBlocks {
		blockHashes[i] = heightBlocks[i].BlockHash[:]
	}
	/*blocks, err := item.GetBlocks(blockHashes)
	if err != nil {
		return nil, fmt.Errorf("error getting db blocks; %w", err)
	}*/
	var blockInfos = make([]*network_pb.BlockInfo, len(heightBlocks))
	for i := range heightBlocks {
		/*txCount, err := GetBlockTxCount(heightBlocks[i].BlockHash)
		if err != nil {
			return nil, fmt.Errorf("error getting block tx count; %w", err)
		}*/
		/*var header []byte
		for _, block := range blocks {
			if bytes.Equal(block.Hash, heightBlocks[i].BlockHash) {
				header = block.Raw
				break
			}
		}*/
		blockInfos[i] = &network_pb.BlockInfo{
			Hash:   heightBlocks[i].BlockHash[:],
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
	if heightBlocks, err := chain.GetHeightBlocksAll(req.Start, req.Wait); err != nil {
		return nil, fmt.Errorf("error getting height blocks all; %w", err)
	} else {
		response.Blocks = make([]*network_pb.BlockHeight, len(heightBlocks))
		for i := range heightBlocks {
			response.Blocks[i] = &network_pb.BlockHeight{
				Height: heightBlocks[i].Height,
				Hash:   heightBlocks[i].BlockHash[:],
			}
		}
	}
	return response, nil
}

func GetBlockTxCount(blockHash [32]byte) (int64, error) {
	blockInfo, err := chain.GetBlockInfo(blockHash)
	if err != nil {
		return 0, fmt.Errorf("error getting block txes; %w", err)
	}
	return int64(blockInfo.TxCount), nil
}

func (s *Server) GetBlockByHash(_ context.Context, req *network_pb.BlockHashRequest) (*network_pb.BlockInfo, error) {
	blockHash, err := chainhash.NewHash(req.GetHash())
	if err != nil {
		return nil, fmt.Errorf("error getting block hash for network server block by hash; %w", err)
	}
	blockHeight, err := chain.GetBlockHeight(*blockHash)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return nil, fmt.Errorf("error getting block height by hash; %w", err)
	}
	var height int64
	if blockHeight != nil {
		height = blockHeight.Height
	}
	txCount, err := GetBlockTxCount(*blockHash)
	if err != nil {
		return nil, fmt.Errorf("error getting block tx count; %w", err)
	}
	block, err := chain.GetBlock(*blockHash)
	if err != nil {
		return nil, fmt.Errorf("error getting block; %w", err)
	}
	return &network_pb.BlockInfo{
		Hash:   blockHash[:],
		Height: height,
		Txs:    txCount,
		Header: block.Raw,
	}, nil
}

func (s *Server) GetBlockByHeight(_ context.Context, req *network_pb.BlockRequest) (*network_pb.BlockInfo, error) {
	heightBlock, err := chain.GetHeightBlockSingle(req.GetHeight())
	if err != nil {
		return nil, fmt.Errorf("error getting height block by height; %w", err)
	}
	block, err := chain.GetBlock(heightBlock.BlockHash)
	if err != nil {
		return nil, fmt.Errorf("error getting block; %w", err)
	}
	return &network_pb.BlockInfo{
		Hash:   heightBlock.BlockHash[:],
		Height: heightBlock.Height,
		Header: block.Raw,
	}, nil
}

func (s *Server) GetUtxos(_ context.Context, req *network_pb.UtxosRequest) (*network_pb.UtxosResponse, error) {
	var utxos []*network_pb.Output
	// TODO: Reimplement if needed
	return &network_pb.UtxosResponse{
		Outputs: utxos,
	}, nil
}

func (s *Server) Run() error {
	if err := s.Start(); err != nil {
		return fmt.Errorf("error starting network server; %w", err)
	}
	// Serve always returns an error
	return fmt.Errorf("error serving network server; %w", s.Serve())
}

func (s *Server) Start() error {
	var err error
	if s.listener, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.Port)); err != nil {
		return fmt.Errorf("failed to listen; %w", err)
	}
	s.grpc = grpc.NewServer()
	network_pb.RegisterNetworkServer(s.grpc, s)
	reflection.Register(s.grpc)
	return nil
}

func (s *Server) Serve() error {
	if err := s.grpc.Serve(s.listener); err != nil {
		return fmt.Errorf("failed to serve; %w", err)
	}
	return fmt.Errorf("network rpc server disconnected")
}

func NewServer(verbose bool, port int) *Server {
	return &Server{
		Verbose: verbose,
		Port:    port,
	}
}
