package run

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/db/server"
)

type Queue struct {
	Port   uint
	Shard  uint
	Server *server.Server
	Error  error
}

func (q *Queue) Start() error {
	q.Server = server.NewServer(q.Port, q.Shard)
	jlog.Logf("Starting queue server on port: %d\n", q.Port)
	go func() {
		if err := q.Server.Run(); !q.Server.Stopped {
			q.Error = jerr.Get("error queue server ended", err)
		}
	}()
	return nil
}

func (q *Queue) End() {
	if q.Server != nil {
		q.Server.Stop()
	}
}

func NewQueue(port uint, shard uint) *Queue {
	return &Queue{
		Port:  port,
		Shard: shard,
	}
}
