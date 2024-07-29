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

type OptionLimit struct {
	Limit int
}

func (o *OptionLimit) Apply(r *queue_pb.RequestPrefixes) {
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

func NewOptionPrefixLimit(limit int) *OptionPrefixLimit {
	return &OptionPrefixLimit{
		Limit: limit,
	}
}

func OptionSinglePrefixLimit() *OptionPrefixLimit {
	return NewOptionPrefixLimit(1)
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

func NewOptionOrder(desc bool) *OptionOrder {
	return &OptionOrder{
		Desc: desc,
	}
}

func OptionNewest() *OptionOrder {
	return NewOptionOrder(true)
}
