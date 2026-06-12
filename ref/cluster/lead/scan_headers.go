package lead

import (
	"context"
	"fmt"
	"log"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/peer"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/conn"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type ScanHeaders struct {
	height   int64
	lastHash *chainhash.Hash
	synced   bool
	peer     *peer.Peer
	Rescan   bool
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
			s.lastHash = &blockHash
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
	if !s.synced {
		return fmt.Errorf("error scan headers peer disconnected before sync complete (height: %d)", s.height)
	}
	log.Printf("ScanHeaders complete at height: %s\n", jfmt.AddCommas(s.height))
	return nil
}

func (s *ScanHeaders) OnVerAck(_ *peer.Peer, _ *wire.MsgVerAck) {
	if s.lastHash == nil {
		s.requestHeadersFrom(wallet.GetGenesisBlock().Hash)
		return
	}
	s.requestLocatorHeaders()
}

func (s *ScanHeaders) OnHeaders(_ *peer.Peer, msg *wire.MsgHeaders) {
	if len(msg.Headers) == 0 {
		s.synced = true
		s.peer.Disconnect()
		return
	}
	ctx := context.TODO()
	var objects []db.Object
	for _, blockHeader := range msg.Headers {
		blockHash := blockHeader.BlockHash()
		if s.lastHash != nil && blockHeader.PrevBlock == *s.lastHash {
			s.height++
		} else {
			height, ok, err := s.resolveHeight(ctx, blockHeader.PrevBlock)
			if err != nil {
				log.Fatalf("error determining header height; %v", err)
			}
			if !ok {
				log.Printf("ScanHeaders header %s does not connect (prev %s); re-requesting from height %s\n",
					blockHash, blockHeader.PrevBlock, jfmt.AddCommas(s.height))
				if err := s.saveHeaders(objects); err != nil {
					log.Fatalf("error saving header objects; %v", err)
				}
				s.requestLocatorHeaders()
				return
			}
			reorged, err := s.clearReorg(ctx, height, blockHash)
			if err != nil {
				log.Fatalf("error clearing reorged height index at height %d; %v", height, err)
			}
			if reorged {
				if err := rollBackSyncStatus(ctx, height-1); err != nil {
					log.Fatalf("error rolling back sync status; %v", err)
				}
			}
			s.height = height
		}
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
		lastHash := blockHash
		s.lastHash = &lastHash
	}
	if err := s.saveHeaders(objects); err != nil {
		log.Fatalf("error saving header objects; %v", err)
	}
	if s.height%2000 == 0 {
		log.Printf("Scanned headers to height %s\n", jfmt.AddCommas(s.height))
	}
	s.requestHeadersFrom(s.lastHash)
}

func (s *ScanHeaders) saveHeaders(objects []db.Object) error {
	if len(objects) == 0 {
		return nil
	}
	if err := db.Save(objects); err != nil {
		return fmt.Errorf("error saving header objects; %w", err)
	}
	return nil
}

func (s *ScanHeaders) resolveHeight(ctx context.Context, prevBlock chainhash.Hash) (int64, bool, error) {
	if prevBlock == *wallet.GetGenesisBlock().Hash {
		return 1, true, nil
	}
	blockHeight, err := chain.GetBlockHeight(ctx, prevBlock)
	if err != nil {
		if client.IsEntryNotFoundError(err) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("error getting previous block height; %w", err)
	}
	return blockHeight.Height + 1, true, nil
}

func (s *ScanHeaders) clearReorg(ctx context.Context, height int64, newHash chainhash.Hash) (bool, error) {
	existing, err := chain.GetHeightBlock(ctx, height)
	if err != nil {
		return false, fmt.Errorf("error getting height block; %w", err)
	}
	var isReorg bool
	for _, hb := range existing {
		if chainhash.Hash(hb.BlockHash) != newHash {
			isReorg = true
			break
		}
	}
	if !isReorg {
		return false, nil
	}
	log.Printf("ScanHeaders reorg at height %s: replacing orphaned branch with %s\n",
		jfmt.AddCommas(height), newHash)
	nextHeight := height
	for {
		heightBlocks, err := chain.GetHeightBlocksAllLimit(ctx, nextHeight, client.HugeLimit, false)
		if err != nil {
			return false, fmt.Errorf("error getting orphaned height blocks; %w", err)
		}
		if len(heightBlocks) == 0 {
			break
		}
		var objects []db.Object
		for _, hb := range heightBlocks {
			objects = append(objects,
				&chain.HeightBlock{Height: hb.Height, BlockHash: hb.BlockHash},
				&chain.BlockHeight{Height: hb.Height, BlockHash: hb.BlockHash},
			)
			nextHeight = hb.Height + 1
		}
		if err := db.Remove(objects); err != nil {
			return false, fmt.Errorf("error removing orphaned height index; %w", err)
		}
	}
	return true, nil
}

func (s *ScanHeaders) requestHeadersFrom(locatorHash *chainhash.Hash) {
	if locatorHash == nil {
		locatorHash = wallet.GetGenesisBlock().Hash
	}
	msgGetHeaders := wire.NewMsgGetHeaders()
	msgGetHeaders.BlockLocatorHashes = append(msgGetHeaders.BlockLocatorHashes, locatorHash)
	s.peer.QueueMessage(msgGetHeaders, nil)
}

func (s *ScanHeaders) requestLocatorHeaders() {
	locatorHashes, err := getBlockLocator(context.TODO(), s.height)
	if err != nil {
		log.Fatalf("error building block locator for scan headers (height: %d); %v", s.height, err)
	}
	msgGetHeaders := wire.NewMsgGetHeaders()
	msgGetHeaders.BlockLocatorHashes = locatorHashes
	s.peer.QueueMessage(msgGetHeaders, nil)
}

func rollBackSyncStatus(ctx context.Context, height int64) error {
	if height < 0 {
		height = 0
	}
	syncStatus, err := item.GetSyncStatus(ctx, item.SyncStatusBlockHeight)
	if err != nil {
		if client.IsEntryNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("error getting sync status; %w", err)
	}
	if syncStatus.Height <= height {
		return nil
	}
	log.Printf("ScanHeaders rolling sync status back from %s to %s due to reorg\n",
		jfmt.AddCommas(syncStatus.Height), jfmt.AddCommas(height))
	if err := db.Save([]db.Object{&item.SyncStatus{
		Name:   item.SyncStatusBlockHeight,
		Height: height,
	}}); err != nil {
		return fmt.Errorf("error saving sync status; %w", err)
	}
	return nil
}

func (s *ScanHeaders) OnVersion(_ *peer.Peer, msg *wire.MsgVersion) {
	log.Printf("ScanHeaders connected to peer: %s (last block: %d)\n", msg.UserAgent, msg.LastBlock)
}

func NewScanHeaders() *ScanHeaders {
	return &ScanHeaders{}
}
