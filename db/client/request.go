package client

import "github.com/memocash/index/db/proto/queue_pb"

type Prefix struct {
	Prefix []byte
	Start  []byte
	Limit  uint32
}

func NewPrefix(prefix []byte) Prefix {
	return Prefix{
		Prefix: prefix,
	}
}

func NewStart(start []byte) Prefix {
	return Prefix{
		Start: start,
	}
}

type Option interface {
	Apply(*queue_pb.RequestPrefixes)
}

type OptionMax struct {
	Max int
}

func (o *OptionMax) Apply(r *queue_pb.RequestPrefixes) {
	r.Max = uint32(o.Max)
}

func NewOptionMax(max int) *OptionMax {
	return &OptionMax{
		Max: max,
	}
}

type OptionNewest struct {
	Newest bool
}

func (o *OptionNewest) Apply(r *queue_pb.RequestPrefixes) {
	r.Newest = o.Newest
}

func NewOptionNewest(newest bool) *OptionNewest {
	return &OptionNewest{
		Newest: newest,
	}
}
