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
	Hash      string `json:"hash"`
	Index     uint32 `json:"index"`
	Timestamp Date   `json:"timestamp"`
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
	Raw       string `json:"raw"`
	Timestamp Date   `json:"timestamp"`
	Height    *int   `json:"height"`
}

type Profile struct {
	Lock      *Lock       `json:"lock"`
	Name      *SetName    `json:"name"`
	Profile   *SetProfile `json:"profile"`
	Pic       *SetPic     `json:"pic"`
}
