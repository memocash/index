package lead

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/peer"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/node/conn"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
)

const MaxHeightBack = 20

type BlockNode struct {
	peer       *peer.Peer
	NewBlock   chan *dbi.Block
	SyncDone   chan struct{}
	synced     bool
	heightBack int64
	lastBlock  *chainhash.Hash
}

func (n *BlockNode) connect() error {
	connection, err := conn.NewConnection(peer.MessageListeners{
		OnVerAck:  n.OnVerAck,
		OnHeaders: n.OnHeaders,
		OnInv:     n.OnInv,
		OnBlock:   n.OnBlock,
		OnReject:  n.OnReject,
		OnPing:    n.OnPing,
		OnVersion: n.OnVersion,
	})
	if err != nil {
		return fmt.Errorf("error getting new outbound peer; %w", err)
	}
	n.peer = connection.Peer
	log.Printf("BlockNode connecting to: %s\n", connection.Address)
	connection.Peer.WaitForDisconnect()
	_ = connection.Net.Close()
	return nil
}

func (n *BlockNode) Start() {
	go func() {
		for {
			n.heightBack = 0
			n.lastBlock = nil
			if err := n.connect(); err != nil {
				log.Fatalf("fatal error connecting block node peer; %v", err)
			}
			log.Printf("BlockNode peer disconnected\n")
			const sleepSeconds = 5
			log.Printf("BlockNode reconnecting to peer after %d seconds\n", sleepSeconds)
			time.Sleep(time.Second * sleepSeconds)
		}
	}()
}

func (n *BlockNode) Stop() {
	if n.peer != nil {
		n.peer.Disconnect()
	}
}

func (n *BlockNode) OnVerAck(_ *peer.Peer, _ *wire.MsgVerAck) {
	msgGetHeaders := wire.NewMsgGetHeaders()
	ctx := context.TODO()
	syncStatus, err := item.GetSyncStatus(ctx, item.SyncStatusBlockHeight)
	if err != nil && !client.IsEntryNotFoundError(err) {
		log.Fatalf("error getting sync status block height; %v", err)
	}
	var height int64
	if syncStatus != nil {
		height = syncStatus.Height
	} else {
		height = int64(config.GetInitBlockHeight())
	}
	var blockHash chainhash.Hash
	heightBlock, err := chain.GetHeightBlockSingle(ctx, height)
	if err != nil {
		if height != 0 || !client.IsEntryNotFoundError(err) {
			log.Fatalf("error getting height block for block node (height: %d); %v", height, err)
		}
		blockHash = *wallet.GetGenesisBlock().Hash
	} else {
		blockHash = heightBlock.BlockHash
	}
	msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, &blockHash)
	n.peer.QueueMessage(msgGetHeaders, nil)
}

func (n *BlockNode) OnHeaders(_ *peer.Peer, msg *wire.MsgHeaders) {
	if len(msg.Headers) == 0 {
		if !n.synced {
			log.Printf("BlockNode sync caught up\n")
			n.synced = true
			n.SyncDone <- struct{}{}
		}
		return
	}
	msgGetData := wire.NewMsgGetData()
	for _, blockHeader := range msg.Headers {
		if n.lastBlock != nil && blockHeader.PrevBlock != *n.lastBlock {
			if blockHeader.BlockHash() == *n.lastBlock {
				continue
			}
			go func() {
				time.Sleep(5 * time.Second)
				n.heightBack++
				if n.heightBack > MaxHeightBack {
					log.Fatalf("error beginning of block loop, potential orphan and height back (%d) "+
						"over max height back (%d)", n.heightBack, MaxHeightBack)
					return
				}
				blockHeight, err := chain.GetBlockHeight(context.TODO(), *n.lastBlock)
				if err != nil {
					log.Fatalf("error getting block height; %v", err)
				}
				heightBlock, err := chain.GetHeightBlockSingle(context.TODO(), blockHeight.Height-n.heightBack)
				if err != nil {
					log.Fatalf("error getting height block; %v", err)
				}
				msgGetHeaders := wire.NewMsgGetHeaders()
				blockHash := chainhash.Hash(heightBlock.BlockHash)
				msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, &blockHash)
				n.peer.QueueMessage(msgGetHeaders, nil)
			}()
			return
		}
		n.heightBack = 0
		err := msgGetData.AddInvVect(&wire.InvVect{
			Type: wire.InvTypeBlock,
			Hash: blockHeader.BlockHash(),
		})
		if err != nil {
			log.Fatalf("error adding block inventory vector from header; %v", err)
		}
	}
	if len(msgGetData.InvList) > 0 {
		n.lastBlock = &msgGetData.InvList[len(msgGetData.InvList)-1].Hash
		n.peer.QueueMessage(msgGetData, nil)
	}
}

func (n *BlockNode) OnInv(_ *peer.Peer, msg *wire.MsgInv) {
	msgGetData := wire.NewMsgGetData()
	for _, invItem := range msg.InvList {
		if invItem.Type == wire.InvTypeBlock {
			err := msgGetData.AddInvVect(&wire.InvVect{
				Type: wire.InvTypeBlock,
				Hash: invItem.Hash,
			})
			if err != nil {
				log.Fatalf("error adding block inventory vector; %v", err)
			}
		}
	}
	if len(msgGetData.InvList) > 0 {
		n.peer.QueueMessage(msgGetData, nil)
	}
}

func (n *BlockNode) OnBlock(_ *peer.Peer, msg *wire.MsgBlock, _ []byte) {
	n.NewBlock <- dbi.WireBlockToBlock(msg)
	blockHash := msg.BlockHash()
	if blockHash.IsEqual(n.lastBlock) {
		msgGetHeaders := wire.NewMsgGetHeaders()
		msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, &blockHash)
		n.peer.QueueMessage(msgGetHeaders, nil)
	}
}

func (n *BlockNode) OnReject(_ *peer.Peer, msg *wire.MsgReject) {
	log.Printf("BlockNode OnReject: %#v\n", msg)
}

func (n *BlockNode) OnPing(_ *peer.Peer, msg *wire.MsgPing) {
	pong := wire.NewMsgPong(msg.Nonce + 1)
	n.peer.QueueMessage(pong, nil)
}

func (n *BlockNode) OnVersion(_ *peer.Peer, msg *wire.MsgVersion) {
	log.Printf("BlockNode connected to peer: %s (last block: %d)\n", msg.UserAgent, msg.LastBlock)
}

func NewBlockNode() *BlockNode {
	return &BlockNode{
		NewBlock: make(chan *dbi.Block),
		SyncDone: make(chan struct{}),
	}
}
