package conn

import (
	"fmt"
	"net"

	"github.com/jchavannes/btcd/peer"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
)

type Connection struct {
	Peer    *peer.Peer
	Net     net.Conn
	Address string
}

func NewConnection(listeners peer.MessageListeners) (*Connection, error) {
	SetBTCDLogLevel()
	address := config.GetNodeHost()
	newPeer, err := peer.NewOutboundPeer(&peer.Config{
		UserAgentName:    "memo-index",
		UserAgentVersion: "0.3.0",
		ChainParams:      wallet.GetMainNetParams(),
		Listeners:        listeners,
	}, address)
	if err != nil {
		return nil, fmt.Errorf("error getting new outbound peer; %w", err)
	}
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("error getting network connection; %w", err)
	}
	newPeer.AssociateConnection(conn)
	return &Connection{
		Peer:    newPeer,
		Net:     conn,
		Address: address,
	}, nil
}
