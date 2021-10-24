package memo

type Out struct {
	TxHash   []byte
	Index    uint32
	Value    int64
	PkScript []byte
}

type SlpOut struct {
	Out
	Token    []byte
	Quantity uint64
	PkHash   []byte
}
