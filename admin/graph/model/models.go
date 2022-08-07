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
	LockHash string      `json:"lock_hash"`
	Name     *SetName    `json:"name"`
	Profile  *SetProfile `json:"profile"`
	Pic      *SetPic     `json:"pic"`
}

type Follow struct {
	TxHash         string `json:"tx_hash"`
	LockHash       string `json:"lock_hash"`
	FollowLockHash string `json:"follow_lock_hash"`
	Unfollow       bool   `json:"unfollow"`
}

type SetName struct {
	TxHash   string `json:"tx_hash"`
	LockHash string `json:"lock_hash"`
	Name     string `json:"name"`
}

type SetPic struct {
	TxHash   string `json:"tx_hash"`
	LockHash string `json:"lock_hash"`
	Pic      string `json:"pic"`
}

type SetProfile struct {
	TxHash   string `json:"tx_hash"`
	LockHash string `json:"lock_hash"`
	Text     string `json:"text"`
}

type Post struct {
	TxHash       string `json:"tx_hash"`
	LockHash     string `json:"lock_hash"`
	Text         string `json:"text"`
	ParentTxHash string `json:"parent_tx_hash"`
}

type Like struct {
	TxHash     string `json:"tx_hash"`
	LockHash   string `json:"lock_hash"`
	PostTxHash string `json:"post_tx_hash"`
	Tip        int64  `json:"tip"`
}
