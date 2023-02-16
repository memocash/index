package run

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/server"
)

type Queue struct {
	Port   int
	Shard  uint
	Server *server.Server
	Error  error
}

func (q *Queue) Start() error {
	q.Server = server.NewServer(q.Port, q.Shard)
	jlog.Logf("Starting test queue server shard %d on port: %d\n", q.Shard, q.Port)
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

func NewQueue(port int, shard uint) *Queue {
	return &Queue{
		Port:  port,
		Shard: shard,
	}
}
