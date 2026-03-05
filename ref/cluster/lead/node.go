package lead

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/node/peer"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"log"
	"time"
)

type Node struct {
	Off      bool
	Peer     *peer.Peer
	NewBlock chan *dbi.Block
	SyncDone chan struct{}
	Verbose  bool
}

func (n *Node) SaveTxs(ctx context.Context, b *dbi.Block) error {
	if n.Off {
		return nil
	}
	n.NewBlock <- b
	return nil
}

func (n *Node) SaveBlock(dbi.BlockInfo) error {
	if n.Off {
		return nil
	}
	return nil
}

func (n *Node) GetBlock(heightBack int64) (*chainhash.Hash, error) {
	if n.Off {
		return nil, nil
	}
	ctx := context.TODO()
	syncStatus, err := item.GetSyncStatus(ctx, item.SyncStatusBlockHeight)
	if err != nil && !client.IsEntryNotFoundError(err) {
		return nil, fmt.Errorf("error getting sync status block height; %w", err)
	}
	var height int64
	if syncStatus != nil {
		height = syncStatus.Height - heightBack
	} else {
		height = int64(config.GetInitBlockHeight()) - 1
	}
	heightBlock, err := chain.GetHeightBlockSingle(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("error getting height block for lead node (height: %d); %w", height, err)
	}
	blockHash := chainhash.Hash(heightBlock.BlockHash)
	return &blockHash, nil
}

func (n *Node) Start(memPool, syncDone bool) {
	go func() {
		for {
			n.Peer = peer.NewConnection(n, n)
			n.Peer.SyncDone = syncDone
			n.Peer.Mempool = memPool
			n.Off = false
			if err := n.Peer.Connect(); err != nil {
				log.Fatalf("fatal error connecting to peer; %v", err)
			}
			log.Printf("node peer disconnected\n")
			n.Off = true
			if n.Peer.SyncDone {
				n.SyncDone <- struct{}{}
				break
			}
			const sleepSeconds = 5
			log.Printf("reconnecting node peer after %d seconds\n", sleepSeconds)
			time.Sleep(time.Second * sleepSeconds)
		}
	}()
}

func (n *Node) Stop() {
	n.Off = true
	n.Peer.Disconnect()
}

func NewNode() *Node {
	return &Node{
		NewBlock: make(chan *dbi.Block),
		SyncDone: make(chan struct{}),
	}
}
