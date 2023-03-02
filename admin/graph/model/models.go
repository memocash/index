package model

type Tx struct {
	Hash  string `json:"hash"`
	Index uint32 `json:"index"`
	Raw   string `json:"raw"`
	Seen  Date   `json:"seen"`
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
	Script    string `json:"script"`
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
	Address string `json:"address"`
	Balance int64  `json:"balance"`
}

type Block struct {
	Hash      string `json:"hash"`
	Raw       string `json:"raw"`
	Timestamp Date   `json:"timestamp"`
	Height    *int   `json:"height"`
	Size      int64  `json:"size"`
	TxCount   int    `json:"tx_count"`
}

type Profile struct {
	Address string      `json:"address"`
	Name    *SetName    `json:"name"`
	Profile *SetProfile `json:"profile"`
	Pic     *SetPic     `json:"pic"`
}

type Follow struct {
	TxHash        string `json:"tx_hash"`
	Address       string `json:"address"`
	FollowAddress string `json:"follow_address"`
	Unfollow      bool   `json:"unfollow"`
}

type SetName struct {
	TxHash  string `json:"tx_hash"`
	Address string `json:"address"`
	Name    string `json:"name"`
}

type SetPic struct {
	TxHash  string `json:"tx_hash"`
	Address string `json:"address"`
	Pic     string `json:"pic"`
}

type SetProfile struct {
	TxHash  string `json:"tx_hash"`
	Address string `json:"address"`
	Text    string `json:"text"`
}

type Post struct {
	TxHash  string `json:"tx_hash"`
	Address string `json:"address"`
	Text    string `json:"text"`
}

type Like struct {
	TxHash     string `json:"tx_hash"`
	Address    string `json:"address"`
	PostTxHash string `json:"post_tx_hash"`
	Tip        int64  `json:"tip"`
}

type Room struct {
	Name string `json:"name"`
}

type RoomFollow struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Unfollow bool   `json:"unfollow"`
	TxHash   string `json:"tx_hash"`
}

type SlpBaton struct {
	Hash      string `json:"hash"`
	Index     uint32 `json:"index"`
	TokenHash string `json:"token_hash"`
}

type SlpGenesis struct {
	Hash       string `json:"hash"`
	TokenType  Uint8  `json:"token_type"`
	Decimals   Uint8  `json:"decimals"`
	BatonIndex uint32 `json:"baton_index"`
	Ticker     string `json:"ticker"`
	Name       string `json:"name"`
	DocURL     string `json:"doc_url"`
	DocHash    string `json:"doc_hash"`
}

type SlpOutput struct {
	Hash      string `json:"hash"`
	Index     uint32 `json:"index"`
	TokenHash string `json:"token_hash"`
	Amount    uint64 `json:"amount"`
}
