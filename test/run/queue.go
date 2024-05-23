package run

import (
	"fmt"
	"github.com/memocash/index/db/server"
	"log"
)

type Queue struct {
	Port   int
	Shard  uint
	Server *server.Server
	Error  error
}

func (q *Queue) Start() error {
	q.Server = server.NewServer(q.Port, q.Shard)
	log.Printf("Starting test queue server shard %d on port: %d\n", q.Shard, q.Port)
	go func() {
		if err := q.Server.Run(); !q.Server.Stopped {
			q.Error = fmt.Errorf("error queue server ended; %w", err)
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
