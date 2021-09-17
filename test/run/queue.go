package run

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/queue"
)

type Queue struct {
	Port   uint
	Server *queue.Server
	Error  error
}

func (q *Queue) Start() error {
	q.Server = queue.NewServer(q.Port)
	go func() {
		err := q.Server.Run()
		q.Error = jerr.Get("error queue server ended", err)
	}()
	return nil
}

func (q *Queue) End() {
	q.Server.Stop()
}

func NewQueue(port uint) *Queue {
	return &Queue{
		Port: port,
	}
}
