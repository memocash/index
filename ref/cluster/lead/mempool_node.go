package lead

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jchavannes/btcd/peer"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/memocash/index/node/conn"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/dbi"
)

type MempoolNode struct {
	peer     *peer.Peer
	NewBlock chan *dbi.Block
}

func (n *MempoolNode) connect() error {
	connection, err := conn.NewConnection(peer.MessageListeners{
		OnVerAck:  n.OnVerAck,
		OnInv:     n.OnInv,
		OnTx:      n.OnTx,
		OnReject:  n.OnReject,
		OnPing:    n.OnPing,
		OnVersion: n.OnVersion,
	})
	if err != nil {
		return fmt.Errorf("error getting new outbound peer; %w", err)
	}
	n.peer = connection.Peer
	log.Printf("%s connecting to: %s\n", NameMempoolNode, connection.Address)
	connection.Peer.WaitForDisconnect()
	_ = connection.Net.Close()
	return nil
}

func (n *MempoolNode) Start() {
	go func() {
		for {
			if err := n.connect(); err != nil {
				log.Fatalf("fatal error connecting mempool node peer; %v", err)
			}
			log.Printf("%s peer disconnected\n", NameMempoolNode)
			const sleepSeconds = 5
			log.Printf("%s reconnecting to peer after %d seconds\n", NameMempoolNode, sleepSeconds)
			time.Sleep(time.Second * sleepSeconds)
		}
	}()
}

func (n *MempoolNode) OnVerAck(_ *peer.Peer, _ *wire.MsgVerAck) {
	n.peer.QueueMessage(wire.NewMsgMemPool(), nil)
}

func (n *MempoolNode) OnInv(_ *peer.Peer, msg *wire.MsgInv) {
	msgGetData := wire.NewMsgGetData()
	for _, invItem := range msg.InvList {
		if invItem.Type == wire.InvTypeTx {
			err := msgGetData.AddInvVect(&wire.InvVect{
				Type: wire.InvTypeTx,
				Hash: invItem.Hash,
			})
			if err != nil {
				log.Fatalf("error adding tx inventory vector; %v", err)
			}
		}
	}
	if len(msgGetData.InvList) > 0 {
		n.peer.QueueMessage(msgGetData, nil)
	}
}

func (n *MempoolNode) OnTx(_ *peer.Peer, msg *wire.MsgTx) {
	log.Printf("%s OnTx: %s, in: %s, out: %s, size: %s\n", NameMempoolNode, msg.TxHash().String(),
		jfmt.AddCommasInt(len(msg.TxIn)), jfmt.AddCommasInt(len(msg.TxOut)), jfmt.AddCommasInt(msg.SerializeSize()))
	n.NewBlock <- dbi.WireBlockToBlock(memo.GetBlockFromTxs([]*wire.MsgTx{msg}, nil))
}

func (n *MempoolNode) OnReject(_ *peer.Peer, msg *wire.MsgReject) {
	log.Printf("%s OnReject: %#v\n", NameMempoolNode, msg)
}

func (n *MempoolNode) OnPing(_ *peer.Peer, msg *wire.MsgPing) {
	pong := wire.NewMsgPong(msg.Nonce + 1)
	n.peer.QueueMessage(pong, nil)
}

func (n *MempoolNode) OnVersion(_ *peer.Peer, msg *wire.MsgVersion) {
	log.Printf("%s connected to peer: %s (last block: %d)\n", NameMempoolNode, msg.UserAgent, msg.LastBlock)
}

func (n *MempoolNode) BroadcastTx(ctx context.Context, msgTx *wire.MsgTx) error {
	var done = make(chan struct{})
	n.peer.QueueMessage(msgTx, done)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("error context timeout")
	}
}

func NewMempoolNode() *MempoolNode {
	return &MempoolNode{
		NewBlock: make(chan *dbi.Block),
	}
}
