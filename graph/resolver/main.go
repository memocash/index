package resolver

import (
	"context"
	"github.com/memocash/index/db/metric"
	"log"
	"time"
)

type InternalError struct {
	err error
}

func (e InternalError) Error() string {
	return e.err.Error()
}

func (e InternalError) Unwrap() error {
	return e.err
}

type Request struct {
	Start time.Time
	Ip    string
	Query string
}

func NewRequest(ip string) *Request {
	return &Request{
		Start: time.Now(),
		Ip:    ip,
	}
}

func (r *Request) Log(urlPlus string) {
	log.Printf("%s %s\n", r.Ip, urlPlus)
}

func (r *Request) GetDuration() time.Duration {
	return time.Since(r.Start)
}

const RequestContextKey = "requestContextKey"

func AttachRequestToContext(ctx context.Context, reqLog *Request) context.Context {
	return context.WithValue(ctx, RequestContextKey, reqLog)
}

func SetContextRequestQuery(ctx context.Context, query string) {
	if r, ok := ctx.Value(RequestContextKey).(*Request); ok {
		r.Query = query
	}
}

func LogContextRequest(ctx context.Context, message string) {
	if r, ok := ctx.Value(RequestContextKey).(*Request); ok {
		r.Log(message)
	}
}

func SetEndPoint(ctx context.Context, endPoint string) {
	metric.AddGraphQuery(endPoint)
	SetContextRequestQuery(ctx, endPoint)
}
