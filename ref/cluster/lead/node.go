package lead

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/node/peer"
	"github.com/memocash/index/ref/dbi"
)

type Node struct {
	Off      bool
	Peer     *peer.Peer
	NewBlock chan *dbi.Block
	Verbose  bool
}

func (n *Node) SaveTxs(block *dbi.Block) error {
	if n.Off {
		return nil
	}
	n.NewBlock <- block
	return nil
}

func (n *Node) SaveBlock(header wire.BlockHeader) error {
	if n.Off {
		return nil
	}
	if err := saver.NewBlock(n.Verbose).SaveBlock(header); err != nil {
		return jerr.Get("error saving block for lead node", err)
	}
	return nil
}

func (n *Node) GetBlock(height int64) ([]byte, error) {
	if n.Off {
		return nil, nil
	}
	hash, err := saver.NewBlock(n.Verbose).GetBlock(height)
	if err != nil {
		return nil, jerr.Get("error getting block for lead node", err)
	}
	return hash, nil
}

func (n *Node) Start() {
	n.Peer = peer.NewConnection(n, n)
	go func() {
		if err := n.Peer.Connect(); err != nil {
			jerr.Get("fatal error connecting to peer", err).Fatal()
		}
		n.Stop()
	}()
}

func (n *Node) Stop() {
	n.Off = true
	n.Peer.Disconnect()
}

func NewNode() *Node {
	return &Node{
		NewBlock: make(chan *dbi.Block),
	}
}
