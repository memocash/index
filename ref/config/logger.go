package config

import (
	"github.com/jchavannes/jgo/jlog"
	"log"
)

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	length := len(bytes)
	jlog.Logf(string(bytes))
	return length, nil
}

func SetLogger() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}
