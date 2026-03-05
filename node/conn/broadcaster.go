package conn

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jchavannes/btcd/peer"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/btclog"
)

type Broadcaster struct {
	peer *peer.Peer
}

func (p *Broadcaster) Connect() error {
	conn, err := NewConnection(peer.MessageListeners{
		OnPing:    p.OnPing,
		OnVersion: p.OnVersion,
	})
	if err != nil {
		return fmt.Errorf("error getting new outbound peer; %w", err)
	}
	p.peer = conn.Peer
	log.Printf("Starting broadcaster node: %s\n", conn.Address)
	conn.Peer.WaitForDisconnect()
	_ = conn.Net.Close()
	return nil
}

func (p *Broadcaster) Disconnect() {
	if p != nil && p.peer != nil {
		p.peer.Disconnect()
	}
}

func (p *Broadcaster) OnPing(_ *peer.Peer, msg *wire.MsgPing) {
	log.Printf("OnPing: %d\n", msg.Nonce)
	pong := wire.NewMsgPong(msg.Nonce + 1)
	p.peer.QueueMessage(pong, nil)
}

func (p *Broadcaster) OnVersion(_ *peer.Peer, msg *wire.MsgVersion) {
	log.Printf("OnVersion: %s (last: %d)\n", msg.UserAgent, msg.LastBlock)
}

func (p *Broadcaster) BroadcastTx(ctx context.Context, msgTx *wire.MsgTx) error {
	var done = make(chan struct{})
	p.peer.QueueMessage(msgTx, done)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("error context timeout")
	}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{}
}

func SetBTCDLogLevel() {
	logger := btclog.NewBackend(os.Stdout).Logger("MEMO")
	logger.SetLevel(btclog.LevelError)
	peer.UseLogger(logger)
}
