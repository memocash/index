package lead

import (
	"context"
	"fmt"
	"log"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/peer"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/conn"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type ScanHeaders struct {
	height    int64
	startHash *chainhash.Hash
	peer      *peer.Peer
	Rescan    bool
}

func (s *ScanHeaders) Run() error {
	if !s.Rescan {
		recentBlock, err := chain.GetRecentHeightBlock(context.Background())
		if err != nil {
			return fmt.Errorf("error getting recent height block; %w", err)
		}
		if recentBlock != nil {
			s.height = recentBlock.Height
			blockHash := chainhash.Hash(recentBlock.BlockHash)
			s.startHash = &blockHash
			log.Printf("ScanHeaders resuming from height: %s\n", jfmt.AddCommas(s.height))
		}
	}
	connection, err := conn.NewConnection(peer.MessageListeners{
		OnVerAck:  s.OnVerAck,
		OnHeaders: s.OnHeaders,
		OnVersion: s.OnVersion,
	})
	if err != nil {
		return fmt.Errorf("error getting new outbound peer; %w", err)
	}
	s.peer = connection.Peer
	log.Printf("ScanHeaders connecting to: %s\n", connection.Address)
	connection.Peer.WaitForDisconnect()
	_ = connection.Net.Close()
	log.Printf("ScanHeaders complete at height: %s\n", jfmt.AddCommas(s.height))
	return nil
}

func (s *ScanHeaders) OnVerAck(_ *peer.Peer, _ *wire.MsgVerAck) {
	var locatorHash *chainhash.Hash
	if s.startHash != nil {
		locatorHash = s.startHash
	} else {
		locatorHash = wallet.GetGenesisBlock().Hash
	}
	msgGetHeaders := wire.NewMsgGetHeaders()
	msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, locatorHash)
	s.peer.QueueMessage(msgGetHeaders, nil)
}

func (s *ScanHeaders) OnHeaders(_ *peer.Peer, msg *wire.MsgHeaders) {
	if len(msg.Headers) == 0 {
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
	log.Printf("ScanHeaders connected to peer: %s (last block: %d)\n", msg.UserAgent, msg.LastBlock)
}

func NewScanHeaders() *ScanHeaders {
	return &ScanHeaders{}
}
