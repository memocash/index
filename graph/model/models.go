package model

type Tx struct {
	Hash     Hash        `json:"hash"`
	Raw      Bytes       `json:"raw"`
	Seen     Date        `json:"seen"`
	Version  int32       `json:"version"`
	LockTime uint32      `json:"locktime"`
	Inputs   []*TxInput  `json:"inputs"`
	Outputs  []*TxOutput `json:"outputs"`
	Blocks   []*TxBlock  `json:"blocks"`
}

type TxOutput struct {
	Hash     Hash       `json:"hash"`
	Index    uint32     `json:"index"`
	Amount   int64      `json:"amount"`
	Script   Bytes      `json:"script"`
	Tx       *Tx        `json:"tx"`
	Lock     *Lock      `json:"lock"`
	Spends   []*TxInput `json:"spends"`
	Slp      *SlpOutput `json:"slp"`
	SlpBaton *SlpBaton  `json:"slp_baton"`
}

type TxInput struct {
	Hash      Hash      `json:"hash"`
	Index     uint32    `json:"index"`
	PrevHash  Hash      `json:"prev_hash"`
	PrevIndex uint32    `json:"prev_index"`
	Script    Bytes     `json:"script"`
	Sequence  uint32    `json:"sequence"`
	Tx        *Tx       `json:"tx"`
	Output    *TxOutput `json:"output"`
}

type Lock struct {
	Address Address  `json:"address"`
	Balance int64    `json:"balance"`
	Profile *Profile `json:"profile"`
	Txs     []*Tx    `json:"txs"`
}

type TxBlock struct {
	TxHash    Hash   `json:"tx_hash"`
	BlockHash Hash   `json:"block_hash"`
	Tx        *Tx    `json:"tx"`
	Block     *Block `json:"block"`
	Index     uint32 `json:"index"`
}

type Block struct {
	Hash      Hash       `json:"hash"`
	Raw       Bytes      `json:"raw"`
	Timestamp Date       `json:"timestamp"`
	Height    *int       `json:"height"`
	Size      int64      `json:"size"`
	TxCount   int        `json:"tx_count"`
	Txs       []*TxBlock `json:"txs"`
}

type Profile struct {
	Address   Address     `json:"address"`
	Name      *SetName    `json:"name"`
	Profile   *SetProfile `json:"profile"`
	Pic       *SetPic     `json:"pic"`
	Lock      *Lock       `json:"lock"`
	Posts     []*Post     `json:"posts"`
	Following []*Follow   `json:"following"`
}

type Follow struct {
	TxHash        Hash    `json:"tx_hash"`
	Address       Address `json:"address"`
	FollowAddress Address `json:"follow_address"`
	Unfollow      bool    `json:"unfollow"`
	Lock          *Lock   `json:"lock"`
	FollowLock    *Lock   `json:"follow_lock"`
	Tx            *Tx     `json:"tx"`
}

type SetName struct {
	TxHash  Hash    `json:"tx_hash"`
	Address Address `json:"address"`
	Name    string  `json:"name"`
}

type SetPic struct {
	TxHash  Hash    `json:"tx_hash"`
	Address Address `json:"address"`
	Pic     string  `json:"pic"`
}

type SetProfile struct {
	TxHash  Hash    `json:"tx_hash"`
	Address Address `json:"address"`
	Text    string  `json:"text"`
}

type Post struct {
	TxHash  Hash    `json:"tx_hash"`
	Address Address `json:"address"`
	Text    string  `json:"text"`
	Lock    *Lock   `json:"lock"`
	Tx      *Tx     `json:"tx"`
	Parent  *Post   `json:"parent"`
	Likes   []*Like `json:"likes"`
	Replies []*Post `json:"replies"`
	Room    *Room   `json:"room"`
}

type Like struct {
	TxHash     Hash    `json:"tx_hash"`
	Address    Address `json:"address"`
	PostTxHash Hash    `json:"post_tx_hash"`
	Tip        int64   `json:"tip"`
	Lock       *Lock   `json:"lock"`
	Tx         *Tx     `json:"tx"`
	Post       *Post   `json:"post"`
}

type Room struct {
	Name      string        `json:"name"`
	Posts     []*Post       `json:"posts"`
	Followers []*RoomFollow `json:"followers"`
}

type RoomFollow struct {
	Name     string  `json:"name"`
	Address  Address `json:"address"`
	Unfollow bool    `json:"unfollow"`
	TxHash   Hash    `json:"tx_hash"`
	Room     *Room   `json:"room"`
	Lock     *Lock   `json:"lock"`
	Tx       *Tx     `json:"tx"`
}

type SlpBaton struct {
	Hash      Hash        `json:"hash"`
	Index     uint32      `json:"index"`
	TokenHash Hash        `json:"token_hash"`
	Genesis   *SlpGenesis `json:"genesis"`
	Output    *TxOutput   `json:"output"`
}

type SlpGenesis struct {
	Hash       Hash       `json:"hash"`
	TokenType  Uint8      `json:"token_type"`
	Decimals   Uint8      `json:"decimals"`
	BatonIndex uint32     `json:"baton_index"`
	Ticker     string     `json:"ticker"`
	Name       string     `json:"name"`
	DocURL     string     `json:"doc_url"`
	DocHash    string     `json:"doc_hash"`
	Tx         *Tx        `json:"tx"`
	Output     *SlpOutput `json:"output"`
	Baton      *SlpBaton  `json:"baton"`
}

type SlpOutput struct {
	Hash      Hash        `json:"hash"`
	Index     uint32      `json:"index"`
	TokenHash Hash        `json:"token_hash"`
	Amount    uint64      `json:"amount"`
	Genesis   *SlpGenesis `json:"genesis"`
	Output    *TxOutput   `json:"output"`
}
