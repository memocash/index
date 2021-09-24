package node

import (
	"encoding/hex"
	"fmt"
	"github.com/jchavannes/bchutil"
	"github.com/jchavannes/btcd/chaincfg"
	"github.com/jchavannes/btcd/peer"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/db/item"
	"net"
)

const (
	DefaultPort = 8333
)

func GetLocalhost() net.IP {
	var localhostByte, _ = hex.DecodeString("00000000000000000000ffff7f000001")
	return localhostByte
}

type Server struct {
	Peer *peer.Peer
	Ip   []byte
	Port uint16
}

func (s *Server) GetAddr() error {
	if s.Peer == nil {
		return jerr.New("error peer not set")
	}
	s.Peer.QueueMessage(wire.NewMsgGetAddr(), nil)
	return nil
}

func (s *Server) Run() error {
	jlog.Logf("IP: %s\n", net.IP(s.Ip))
	params := &chaincfg.MainNetParams
	params.Net = bchutil.MainnetMagic
	var err error
	connectionAddress := fmt.Sprintf("%s:%d", net.IP(s.Ip), s.Port)
	s.Peer, err = peer.NewOutboundPeer(&peer.Config{
		UserAgentName:    "memo-node",
		UserAgentVersion: "0.3.0",
		ChainParams:      params,
		Listeners: peer.MessageListeners{
			OnAddr: func(p *peer.Peer, msg *wire.MsgAddr) {
				jlog.Logf("on addr: %d\n", len(msg.AddrList))
				var objects = make([]item.Object, len(msg.AddrList)*3)
				for i := range msg.AddrList {
					objects[i*3] = &item.Peer{
						Ip:       msg.AddrList[i].IP,
						Port:     msg.AddrList[i].Port,
						Services: uint64(msg.AddrList[i].Services),
					}
					objects[i*3+1] = &item.PeerFound{
						Ip:         msg.AddrList[i].IP,
						Port:       msg.AddrList[i].Port,
						FinderIp:   s.Ip,
						FinderPort: s.Port,
					}
					objects[i*3+2] = &item.FoundPeer{
						Ip:        s.Ip,
						Port:      s.Port,
						FoundIp:   msg.AddrList[i].IP,
						FoundPort: msg.AddrList[i].Port,
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
	}, connectionAddress)
	if err != nil {
		return jerr.Get("error getting new outbound bitcoinPeer", err)
	}
	jlog.Logf("Starting node: %s\n", connectionAddress)
	conn, err := net.Dial("tcp", connectionAddress)
	if err != nil {
		return jerr.Get("error getting network connection", err)
	}
	s.Peer.AssociateConnection(conn)
	s.Peer.WaitForDisconnect()
	return jerr.Newf("error node disconnected")
}

func (s *Server) Disconnect() {
	if s.Peer != nil {
		s.Peer.Disconnect()
	}
}

func NewServer(ip []byte, port uint16) *Server {
	return &Server{
		Ip:   ip,
		Port: port,
	}
}
