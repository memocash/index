package process

import "github.com/memocash/index/node/obj/status"

type Status interface {
	GetHeight() (int64, error)
	SetHeight(int64) error
}

type StatusHeight interface {
	GetHeight() (status.BlockHeight, error)
	SetHeight(status.BlockHeight) error
}
