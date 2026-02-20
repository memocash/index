package maint

import (
	"fmt"
	"github.com/jchavannes/btcd/peer"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	nodePeer "github.com/memocash/index/node/peer"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/config"
	"log"
	"net"
)

type ScanHeaders struct {
	height int64
	peer   *peer.Peer
}

func (s *ScanHeaders) Run() error {
	nodePeer.SetBtcdLogLevel()
	connectionString := config.GetNodeHost()
	newPeer, err := peer.NewOutboundPeer(&peer.Config{
		UserAgentName:    "memo-index",
		UserAgentVersion: "0.3.0",
		ChainParams:      wallet.GetMainNetParams(),
		Listeners: peer.MessageListeners{
			OnVerAck:  s.OnVerAck,
			OnHeaders: s.OnHeaders,
			OnVersion: s.OnVersion,
		},
	}, connectionString)
	if err != nil {
		return fmt.Errorf("error getting new outbound peer; %w", err)
	}
	s.peer = newPeer
	log.Printf("Starting header scan, connecting to: %s\n", connectionString)
	conn, err := net.Dial("tcp", connectionString)
	if err != nil {
		return fmt.Errorf("error getting network connection; %w", err)
	}
	newPeer.AssociateConnection(conn)
	newPeer.WaitForDisconnect()
	log.Printf("Header scan complete, saved %s headers\n", jfmt.AddCommas(s.height))
	return nil
}

func (s *ScanHeaders) OnVerAck(_ *peer.Peer, _ *wire.MsgVerAck) {
	genesisHash := wallet.GetGenesisBlock().Hash
	msgGetHeaders := wire.NewMsgGetHeaders()
	msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, genesisHash)
	s.peer.QueueMessage(msgGetHeaders, nil)
}

func (s *ScanHeaders) OnHeaders(_ *peer.Peer, msg *wire.MsgHeaders) {
	if len(msg.Headers) == 0 {
		log.Printf("No more headers, scan complete at height %s\n", jfmt.AddCommas(s.height))
		s.peer.Disconnect()
		return
	}
	var objects []db.Object
	for _, blockHeader := range msg.Headers {
		s.height++
		blockHash := blockHeader.BlockHash()
		headerRaw := memo.GetRawBlockHeader(*blockHeader)
		objects = append(objects,
			&chain.Block{
				Hash: blockHash,
				Raw:  headerRaw,
			},
			&chain.BlockHeight{
				BlockHash: blockHash,
				Height:    s.height,
			},
			&chain.HeightBlock{
				Height:    s.height,
				BlockHash: blockHash,
			},
		)
	}
	if err := db.Save(objects); err != nil {
		log.Fatalf("error saving header objects; %v", err)
	}
	if s.height%2000 == 0 {
		log.Printf("Scanned headers to height %s\n", jfmt.AddCommas(s.height))
	}
	lastHeader := msg.Headers[len(msg.Headers)-1]
	lastHash := lastHeader.BlockHash()
	msgGetHeaders := wire.NewMsgGetHeaders()
	msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, &lastHash)
	s.peer.QueueMessage(msgGetHeaders, nil)
}

func (s *ScanHeaders) OnVersion(_ *peer.Peer, msg *wire.MsgVersion) {
	log.Printf("Connected to peer: %s (last block: %d)\n", msg.UserAgent, msg.LastBlock)
}

func NewScanHeaders() *ScanHeaders {
	return &ScanHeaders{}
}
