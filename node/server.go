package node

import (
	"github.com/jchavannes/bchutil"
	"github.com/jchavannes/btcd/chaincfg"
	"github.com/jchavannes/btcd/peer"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/db/item"
	"net"
)

var KnownNodes = []string{
	"127.0.0.1:8333",
}

type Server struct {
}

func (s *Server) Run() error {
	//var hasAskedForPeers bool
	params := &chaincfg.MainNetParams
	params.Net = bchutil.MainnetMagic
	var bitcoinPeer *peer.Peer
	var err error
	bitcoinPeer, err = peer.NewOutboundPeer(&peer.Config{
		UserAgentName:    "memo-node",
		UserAgentVersion: "0.3.0",
		ChainParams:      params,
		Listeners: peer.MessageListeners{
			OnAddr: func(p *peer.Peer, msg *wire.MsgAddr) {
				jlog.Logf("on addr: %d\n", len(msg.AddrList))
				var objects = make([]item.Object, len(msg.AddrList))
				for i := range msg.AddrList {
					objects[i] = &item.Peer{
						Ip:       msg.AddrList[i].IP,
						Port:     msg.AddrList[i].Port,
						Services: uint64(msg.AddrList[i].Services),
					}
				}
				if err := item.Save(objects); err != nil {
					jerr.Get("error saving peers", err).Print()
				}
			},
			OnVerAck: func(p *peer.Peer, msg *wire.MsgVerAck) {
				jlog.Log("ver ack from peer")
			},
			OnHeaders: func(p *peer.Peer, msg *wire.MsgHeaders) {
				jlog.Log("on headers from peer")
				for _, header := range msg.Headers {
					jlog.Logf("header: %s %s\n", header.BlockHash(), header.Timestamp.Format("2006-01-02 15:04:05"))
				}
			},
			/*OnInv: func(p *peer.Peer, msg *wire.MsgInv) {
				jlog.Log("on inv from peer", msg.InvList)
				for _, inv := range msg.InvList {
					jlog.Logf("inv type: %s, hash: %s\n", inv.Type, inv.Hash)
				}
			},*/
			OnBlock: func(p *peer.Peer, msg *wire.MsgBlock, buf []byte) {
				jlog.Log("on block from peer")
				jlog.Logf("block: %s, txs: %d\n", msg.BlockHash(), len(msg.Transactions))
			},
			OnTx: func(p *peer.Peer, msg *wire.MsgTx) {
				jlog.Log("on tx from peer")
				jlog.Logf("tx: %s\n", msg.TxHash())
			},
			OnReject: func(p *peer.Peer, msg *wire.MsgReject) {
				jlog.Log("on reject from peer")
				jlog.Logf("msg: %v\n", msg)
			},
			OnPing: func(p *peer.Peer, msg *wire.MsgPing) {
				jlog.Log("on ping from peer")
				jlog.Logf("Nonce: %d\n", msg.Nonce)
				/*if !hasAskedForPeers {
					bitcoinPeer.QueueMessage(wire.NewMsgGetAddr(), nil)
					hasAskedForPeers = true
					jlog.Log("queued get addr")
				}*/
			},
			OnMerkleBlock: func(p *peer.Peer, msg *wire.MsgMerkleBlock) {
				jlog.Log("on merkle from peer")
				jlog.Logf("merkle block: %s, txs: %d, hashes: %d\n",
					msg.Header.BlockHash(), msg.Transactions, len(msg.Hashes))
			},
			OnVersion: func(p *peer.Peer, msg *wire.MsgVersion) {
				jlog.Log("on version from peer")
				jlog.Logf("version: %d, user agent: %s\n", msg.ProtocolVersion, msg.UserAgent)
			},
		},
	}, KnownNodes[0])
	if err != nil {
		return jerr.Get("error getting new outbound bitcoinPeer", err)
	}
	jlog.Logf("Starting node: %s\n", KnownNodes[0])
	conn, err := net.Dial("tcp", KnownNodes[0])
	if err != nil {
		return jerr.Get("error getting network connection", err)
	}
	bitcoinPeer.AssociateConnection(conn)
	bitcoinPeer.WaitForDisconnect()
	return nil
}

func NewServer() *Server {
	return &Server{}
}
