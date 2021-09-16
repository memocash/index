package main

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/api"
)

func main() {
	if err := api.NewServer().Run(); err != nil {
		jerr.Get("fatal error running api server", err).Fatal()
	}
}
