package resolver

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/metric"
	"log"
	"strings"
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
	Url   string
	Query string
	Size  int
}

func NewRequest(ip, url string) *Request {
	return &Request{
		Start: time.Now(),
		Ip:    ip,
		Url:   url,
	}
}

func (r *Request) Log(messages ...string) {
	if r.Query != "" {
		messages = append([]string{fmt.Sprintf("(%s)", r.Query)}, messages...)
	}
	log.Printf("%s %s %s\n", r.Ip, r.Url, strings.Join(messages, " "))
}

func (r *Request) GetDuration() time.Duration {
	return time.Since(r.Start)
}

func (r *Request) LogFinal(messages ...string) {
	if r.Size > 0 {
		messages = append(messages, fmt.Sprintf("%.1fkb", float32(r.Size)/1000))
	}
	messages = append(messages, fmt.Sprintf("%dms", r.GetDuration().Milliseconds()))
	r.Log(messages...)
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

func OpenSubscriptionWithRequest(ctx context.Context, endPoint string) {
	if r, ok := ctx.Value(RequestContextKey).(*Request); ok {
		r.Query = "sub:" + endPoint
		r.Log("[open]")
	}
}
