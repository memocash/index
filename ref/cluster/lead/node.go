package lead

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/node/peer"
	"github.com/memocash/index/ref/dbi"
	"time"
)

type Node struct {
	Off      bool
	Peer     *peer.Peer
	NewBlock chan *dbi.Block
	SyncDone chan struct{}
	Verbose  bool
}

func (n *Node) SaveTxs(b *dbi.Block) error {
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
	hash, err := saver.NewBlock(n.Verbose).GetBlock(heightBack + 1)
	if err != nil {
		return nil, jerr.Get("error getting block for lead node", err)
	}
	return hash, nil
}

func (n *Node) Start(memPool, syncDone bool) {
	go func() {
		for {
			n.Peer = peer.NewConnection(n, n)
			n.Peer.SyncDone = syncDone
			n.Peer.Mempool = memPool
			n.Off = false
			if err := n.Peer.Connect(); err != nil {
				jerr.Get("fatal error connecting to peer", err).Fatal()
			}
			jlog.Logf("node peer disconnected\n")
			n.Off = true
			if n.Peer.SyncDone {
				n.SyncDone <- struct{}{}
				break
			}
			const sleepSeconds = 5
			jlog.Logf("reconnecting node peer after %d seconds\n", sleepSeconds)
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
