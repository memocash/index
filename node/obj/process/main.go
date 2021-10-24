package process

type Status interface {
	GetHeight() (int64, error)
	SetHeight(int64) error
}

type StatusHeight interface {
	GetHeight() (BlockHeight, error)
	SetHeight(BlockHeight) error
}

type BlockHeight struct {
	Height int64
	Block  []byte
}
