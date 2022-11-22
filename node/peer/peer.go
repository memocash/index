package peer

import (
	"bytes"
	"context"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/peer"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/btclog"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"net"
	"os"
	"time"
)

const (
	MaxHeightBack = 20
)

type Peer struct {
	peer        *peer.Peer
	HandleError func(error)
	BlockSave   dbi.BlockSave
	TxSave      dbi.TxSave
	LastBlock   *chainhash.Hash
	HasExisting bool
	HeightBack  int64
	SyncDone    bool
}

func (p *Peer) Error(err error) {
	if p.HandleError != nil {
		p.HandleError(err)
	} else {
		jerr.Get("fatal peer node error", err).Fatal()
	}
}

func (p *Peer) Connect() error {
	SetBtcdLogLevel()
	connectionString := config.GetNodeHost()
	newPeer, err := peer.NewOutboundPeer(&peer.Config{
		UserAgentName:    "memo-index",
		UserAgentVersion: "0.3.0",
		ChainParams:      wallet.GetMainNetParams(),
		Listeners: peer.MessageListeners{
			OnVerAck:      p.OnVerAck,
			OnHeaders:     p.OnHeaders,
			OnInv:         p.OnInv,
			OnBlock:       p.OnBlock,
			OnTx:          p.OnTx,
			OnReject:      p.OnReject,
			OnPing:        p.OnPing,
			OnMerkleBlock: p.OnMerkleBlock,
			OnVersion:     p.OnVersion,
		},
	}, connectionString)
	if err != nil {
		return jerr.Get("error getting new outbound peer", err)
	}
	p.peer = newPeer
	jlog.Logf("Starting node: %s\n", connectionString)
	conn, err := net.Dial("tcp", connectionString)
	if err != nil {
		return jerr.Get("error getting network connection", err)
	}
	newPeer.AssociateConnection(conn)
	newPeer.WaitForDisconnect()
	return nil
}

func (p *Peer) Disconnect() {
	p.peer.Disconnect()
}

func (p *Peer) OnVerAck(_ *peer.Peer, _ *wire.MsgVerAck) {
	if p.BlockSave == nil {
		if p.TxSave == nil {
			return
		}
		p.peer.QueueMessage(wire.NewMsgMemPool(), nil)
		return
	}
	msgGetHeaders := wire.NewMsgGetHeaders()
	blockHashByte, err := p.BlockSave.GetBlock(0)
	if err != nil {
		p.Error(jerr.Get("error getting node block", err))
		return
	}
	if len(blockHashByte) > 0 && !bytes.Equal(blockHashByte, wallet.GetGenesisBlock().Hash.CloneBytes()) {
		p.HasExisting = true
		blockHash, err := chainhash.NewHash(blockHashByte)
		if err != nil {
			p.Error(jerr.Get("error getting block hash for get headers", err))
		} else {
			msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, blockHash)
		}
	}
	if len(msgGetHeaders.BlockLocatorHashes) == 0 {
		initBlockParent := config.GetInitBlockParent()
		if len(initBlockParent) == 0 {
			initBlock := config.GetInitBlock()
			if initBlock == "" {
				p.Error(jerr.Newf("error init block not set"))
				return
			}
			p.LastBlock, err = chainhash.NewHashFromStr(initBlock)
			if err != nil {
				p.Error(jerr.Get("error getting init block", err))
				return
			}
			msgGetData := wire.NewMsgGetData()
			err := msgGetData.AddInvVect(&wire.InvVect{
				Type: wire.InvTypeBlock,
				Hash: *p.LastBlock,
			})
			if err != nil {
				p.Error(jerr.Get("error adding init block inventory vector", err))
				return
			}
			p.peer.QueueMessage(msgGetData, nil)
			return
		}
		blockHash, err := chainhash.NewHashFromStr(initBlockParent)
		if err != nil {
			p.Error(jerr.Get("error getting block hash for init block parent", err))
			return
		}
		msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, blockHash)
	}
	p.peer.QueueMessage(msgGetHeaders, nil)
}

func (p *Peer) OnHeaders(_ *peer.Peer, msg *wire.MsgHeaders) {
	if jutil.IsNil(p.BlockSave) {
		return
	}
	if len(msg.Headers) == 0 {
		jlog.Logf("No headers received, disconnecting, sync done: %t\n", p.SyncDone)
		if !p.SyncDone {
			p.SyncDone = true
			p.Disconnect()
		}
		return
	}
	msgGetData := wire.NewMsgGetData()
	for _, blockHeader := range msg.Headers {
		blockHash := blockHeader.BlockHash()
		if p.HasExisting && bytes.Equal(blockHash.CloneBytes(), wallet.GetFirstBlock().Hash.CloneBytes()) {
			go func() {
				time.Sleep(5 * time.Second)
				p.HeightBack++
				if p.HeightBack > MaxHeightBack {
					p.Error(jerr.Newf("error beginning of block loop, potential orphan and height back (%d) "+
						"over max height back (%d)", p.HeightBack, MaxHeightBack))
					return
				}
				blockHashByte, err := p.BlockSave.GetBlock(p.HeightBack)
				if err != nil {
					p.Error(jerr.Get("error getting node block after orphan", err))
					return
				}
				blockHash, err := chainhash.NewHash(blockHashByte)
				if err != nil {
					p.Error(jerr.Get("error getting block hash for get headers for orphan", err))
					return
				}
				msgGetHeaders := wire.NewMsgGetHeaders()
				msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, blockHash)
				p.peer.QueueMessage(msgGetHeaders, nil)
				return
			}()
			//p.Error(jerr.New("error beginning of block loop, potentially due to orphan?"))
			return
		}
		p.HeightBack = 0
		err := msgGetData.AddInvVect(&wire.InvVect{
			Type: wire.InvTypeBlock,
			Hash: blockHeader.BlockHash(),
		})
		if err != nil {
			p.Error(jerr.Get("error adding block inventory vector from header", err))
		}
	}
	if len(msgGetData.InvList) > 0 {
		p.LastBlock = &msgGetData.InvList[len(msgGetData.InvList)-1].Hash
		p.peer.QueueMessage(msgGetData, nil)
	}
}

func (p *Peer) OnInv(_ *peer.Peer, msg *wire.MsgInv) {
	msgGetData := wire.NewMsgGetData()
	for _, invItem := range msg.InvList {
		switch invItem.Type {
		case wire.InvTypeTx:
			if p.BlockSave != nil {
				continue
			}
			err := msgGetData.AddInvVect(&wire.InvVect{
				Type: wire.InvTypeTx,
				Hash: invItem.Hash,
			})
			if err != nil {
				p.Error(jerr.Get("error adding tx inventory vector", err))
			}
		case wire.InvTypeBlock:
			if jutil.IsNil(p.BlockSave) {
				return
			}
			err := msgGetData.AddInvVect(&wire.InvVect{
				Type: wire.InvTypeBlock,
				Hash: invItem.Hash,
			})
			if err != nil {
				p.Error(jerr.Get("error adding block inventory vector", err))
			}
		}
	}
	if len(msgGetData.InvList) > 0 {
		p.peer.QueueMessage(msgGetData, nil)
	}
}

func (p *Peer) OnBlock(_ *peer.Peer, msg *wire.MsgBlock, _ []byte) {
	if p.TxSave != nil {
		err := p.TxSave.SaveTxs(dbi.WireBlockToBlock(msg))
		if err != nil {
			p.Error(jerr.Get("error saving txs", err))
		}
	}
	// Save block second in case exit/failure during saving transactions will requeue block again
	if p.BlockSave != nil {
		err := p.BlockSave.SaveBlock(dbi.BlockInfo{
			Header:  msg.Header,
			Size:    int64(msg.SerializeSize()),
			TxCount: len(msg.Transactions),
		})
		if err != nil {
			p.Error(jerr.Get("error saving block", err))
		}
	}
	blockHash := msg.BlockHash()
	if blockHash.IsEqual(p.LastBlock) {
		msgGetHeaders := wire.NewMsgGetHeaders()
		msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, &blockHash)
		p.peer.QueueMessage(msgGetHeaders, nil)
	}
}

func (p *Peer) OnTx(_ *peer.Peer, msg *wire.MsgTx) {
	if p.TxSave != nil {
		jlog.Logf("OnTx: %s\n", msg.TxHash().String())
		err := p.TxSave.SaveTxs(dbi.WireBlockToBlock(memo.GetBlockFromTxs([]*wire.MsgTx{msg}, nil)))
		if err != nil {
			p.Error(jerr.Get("error saving new tx", err))
		}
		return
	}
}

func (p *Peer) OnReject(_ *peer.Peer, msg *wire.MsgReject) {
	jlog.Logf("OnReject: %#v\n", msg)
}

func (p *Peer) OnPing(_ *peer.Peer, msg *wire.MsgPing) {
	jlog.Logf("OnPing: %d\n", msg.Nonce)
	pong := wire.NewMsgPong(msg.Nonce + 1)
	p.peer.QueueMessage(pong, nil)
}

func (p *Peer) OnMerkleBlock(_ *peer.Peer, msg *wire.MsgMerkleBlock) {
	jlog.Logf("OnMerkleBlock: %#v\n", msg)
}

func (p *Peer) OnVersion(_ *peer.Peer, msg *wire.MsgVersion) {
	jlog.Logf("OnVersion: %s (last: %d)\n", msg.UserAgent, msg.LastBlock)
}

func (p *Peer) BroadcastTx(ctx context.Context, msgTx *wire.MsgTx) error {
	var done = make(chan struct{})
	p.peer.QueueMessage(msgTx, done)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return jerr.Newf("error context timeout")
	}
}

func NewConnection(txSave dbi.TxSave, blockSave dbi.BlockSave) *Peer {
	return &Peer{
		BlockSave: blockSave,
		TxSave:    txSave,
	}
}

func SetBtcdLogLevel() {
	logger := btclog.NewBackend(os.Stdout).Logger("MEMO")
	logger.SetLevel(btclog.LevelError)
	peer.UseLogger(logger)
}
