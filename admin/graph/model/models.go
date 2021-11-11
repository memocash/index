package model

type Tx struct {
	Hash string `json:"hash"`
	Raw  string `json:"raw"`
}

type TxOutput struct {
	Hash   string `json:"hash"`
	Index  uint32 `json:"index"`
	Amount int64  `json:"amount"`
	Script string `json:"script"`
}

type TxInput struct {
	Hash      string `json:"hash"`
	Index     uint32 `json:"index"`
	PrevHash  string `json:"prev_hash"`
	PrevIndex uint32 `json:"prev_index"`
}

type DoubleSpend struct {
	Hash  string `json:"hash"`
	Index uint32 `json:"index"`
}

type TxLost struct {
	Hash string `json:"hash"`
}

type TxSuspect struct {
	Hash string `json:"hash"`
}

type Lock struct {
	Hash    string `json:"hash"`
	Address string `json:"address"`
	Balance int64  `json:"balance"`
}

type Block struct {
	Hash      string `json:"hash"`
	Height    *int   `json:"height"`
}
