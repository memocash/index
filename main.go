package main

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/api"
	"github.com/memocash/server/node"
)

func main() {
	var errorHandler = make(chan error)
	go func() {
		err := api.NewServer().Run()
		errorHandler <- jerr.Get("fatal error running api server", err)
	}()
	go func() {
		err := node.NewServer().Run()
		errorHandler <- jerr.Get("fatal error running node server", err)
	}()
	jerr.Get("fatal memo server error encountered", <-errorHandler).Fatal()
}
