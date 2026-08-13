package client

import (
	"time"

	"github.com/memocash/index/db/proto/queue_pb"
)

type Prefix struct {
	Prefix []byte
	Start  []byte
	Limit  uint32
}

// FilterPattern is one arm of a filtered get: uid and data filters with a
// resume cursor and an optional per-arm result cap. A legacy prefix query is
// the arm {Uid: NewPatternPrefix(prefix)}.
type FilterPattern struct {
	Uid   *Pattern
	Data  *Pattern
	Start []byte // asc: inclusive lower bound; desc: exclusive upper bound
	Limit uint32
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

// Option applies to both read calls; every option implements both methods so
// one constructor set works for GetByPrefixes (deprecated) and GetFiltered.
type Option interface {
	Apply(*queue_pb.RequestPrefixes)
	ApplyFilter(*queue_pb.FilterRequest)
}

type OptionLimit struct {
	Limit int
}

func (o *OptionLimit) Apply(r *queue_pb.RequestPrefixes) {
	r.Limit = uint32(o.Limit)
}

func (o *OptionLimit) ApplyFilter(r *queue_pb.FilterRequest) {
	r.Limit = uint32(o.Limit)
}

func NewOptionLimit(limit int) *OptionLimit {
	return &OptionLimit{
		Limit: limit,
	}
}

func OptionLargeLimit() *OptionLimit {
	return NewOptionLimit(LargeLimit)
}

func OptionExLargeLimit() *OptionLimit {
	return NewOptionLimit(ExLargeLimit)
}

func OptionHugeLimit() *OptionLimit {
	return NewOptionLimit(HugeLimit)
}

type OptionPrefixLimit struct {
	Limit int
}

func (o *OptionPrefixLimit) Apply(r *queue_pb.RequestPrefixes) {
	for i := range r.Prefixes {
		r.Prefixes[i].Limit = uint32(o.Limit)
	}
}

func (o *OptionPrefixLimit) ApplyFilter(r *queue_pb.FilterRequest) {
	for i := range r.Patterns {
		r.Patterns[i].Limit = uint32(o.Limit)
	}
}

func NewOptionPrefixLimit(limit int) *OptionPrefixLimit {
	return &OptionPrefixLimit{
		Limit: limit,
	}
}

func OptionSinglePrefixLimit() *OptionPrefixLimit {
	return NewOptionPrefixLimit(1)
}

// OptionTimeout overrides DefaultGetTimeout for one read call. It is consumed
// client-side (a filtered scan of millions of rows can legitimately outrun
// the 60s default); the Apply methods are no-ops since nothing about it goes
// on the wire.
type OptionTimeout struct {
	Timeout time.Duration
}

func (o *OptionTimeout) Apply(*queue_pb.RequestPrefixes) {}

func (o *OptionTimeout) ApplyFilter(*queue_pb.FilterRequest) {}

func NewOptionTimeout(timeout time.Duration) *OptionTimeout {
	return &OptionTimeout{
		Timeout: timeout,
	}
}

type OptionOrder struct {
	Desc bool
}

func (o *OptionOrder) Apply(r *queue_pb.RequestPrefixes) {
	if o.Desc {
		r.Order = queue_pb.Order_DESC
	} else {
		r.Order = queue_pb.Order_ASC
	}
}

func (o *OptionOrder) ApplyFilter(r *queue_pb.FilterRequest) {
	if o.Desc {
		r.Order = queue_pb.Order_DESC
	} else {
		r.Order = queue_pb.Order_ASC
	}
}

func NewOptionOrder(desc bool) *OptionOrder {
	return &OptionOrder{
		Desc: desc,
	}
}

func OptionNewest() *OptionOrder {
	return NewOptionOrder(true)
}
