package memo

type Out struct {
	TxHash   []byte
	Index    uint32
	Value    int64
	PkScript []byte
	LockHash []byte
}

type SlpOut struct {
	Out
	Token    []byte
	Quantity uint64
	PkHash   []byte
}

type InOut struct {
	Hash      []byte
	Index     uint32
	PrevHash  []byte
	PrevIndex uint32
}
