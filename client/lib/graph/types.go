package graph

import (
	"github.com/memocash/index/ref/bitcoin/wallet"
	"time"
)

type Address struct {
	Address string `json:"address"`
	Txs     []Tx   `json:"txs"`
}

type Output struct {
	Tx       Tx        `json:"tx"`
	Hash     string    `json:"hash"`
	Script   string    `json:"script"`
	Index    int       `json:"index"`
	Amount   int64     `json:"amount"`
	Spends   []Input   `json:"spends"`
	Lock     Lock      `json:"lock"`
	Slp      *Slp      `json:"slp"`
	SlpBaton *SlpBaton `json:"slp_baton"`
}

type Slp struct {
	Hash      string      `json:"hash"`
	Index     uint32      `json:"index"`
	TokenHash string      `json:"token_hash"`
	Amount    uint64      `json:"amount"`
	Genesis   *SlpGenesis `json:"genesis"`
}

type SlpBaton struct {
	Hash      string      `json:"hash"`
	Index     uint32      `json:"index"`
	TokenHash string      `json:"token_hash"`
	Genesis   *SlpGenesis `json:"genesis"`
}

type SlpGenesis struct {
	Hash      string `json:"hash"`
	TokenType uint8  `json:"token_type"`
	Decimals  uint8  `json:"decimals"`
	Ticker    string `json:"ticker"`
	Name      string `json:"name"`
	DocUrl    string `json:"doc_url"`
}

type Input struct {
	Tx        Tx     `json:"tx"`
	Hash      string `json:"hash"`
	Index     int    `json:"index"`
	PrevHash  string `json:"prev_hash"`
	PrevIndex int    `json:"prev_index"`
	Output    Output `json:"output"`
}

type Tx struct {
	Hash    string    `json:"hash"`
	Raw     string    `json:"raw"`
	Seen    time.Time `json:"seen"`
	Inputs  []Input   `json:"inputs"`
	Outputs []Output  `json:"outputs"`
	Blocks  []TxBlock `json:"blocks"`
}

type TxBlock struct {
	TxHash    string `json:"tx_hash"`
	BlockHash string `json:"block_hash"`
	Tx        Tx     `json:"tx"`
	Block     Block  `json:"block"`
	Index     int    `json:"index"`
}

type Block struct {
	Hash      string    `json:"hash"`
	Height    int64     `json:"height"`
	Timestamp time.Time `json:"timestamp"`
}

type Lock struct {
	Address string `json:"address"`
}

type Post struct {
	TxHash  string `json:"tx_hash"`
	Address string `json:"address"`
	Text    string `json:"text"`
	Tx      *Tx    `json:"tx"`
}

// The following are not part of graph but are used in queries

type AddressUpdate struct {
	Address wallet.Addr
	Time    time.Time
}
