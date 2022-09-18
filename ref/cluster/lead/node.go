package lead

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/node/peer"
)

type Node struct {
	Off      bool
	Peer     *peer.Peer
	NewBlock chan *wire.MsgBlock
	Verbose  bool
}

func (n *Node) SaveTxs(block *wire.MsgBlock) error {
	if n.Off {
		return nil
	}
	txRawSaver := saver.NewTxRaw(n.Verbose)
	if err := txRawSaver.SaveTxs(block); err != nil {
		return jerr.Get("error saving raw txs for lead node", err)
	}
	n.NewBlock <- block
	return nil
}

func (n *Node) SaveBlock(header wire.BlockHeader) error {
	if n.Off {
		return nil
	}
	blockSaver := saver.BlockSaver(n.Verbose)
	if err := blockSaver.SaveBlock(header); err != nil {
		return jerr.Get("error saving block for lead node", err)
	}
	return nil
}

func (n *Node) GetBlock(height int64) ([]byte, error) {
	if n.Off {
		return nil, nil
	}
	blockSaver := saver.BlockSaver(n.Verbose)
	hash, err := blockSaver.GetBlock(height)
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
		NewBlock: make(chan *wire.MsgBlock),
	}
}
