//go:build debug

package main

import (
	"github.com/memocash/index/cmd"
	"github.com/memocash/index/cmd/debug"
	"log"
)

func main() {
	if err := cmd.Execute(
		debug.GetCommand(),
	); err != nil {
		log.Fatalf("fatal error executing debug command; %v", err)
	}
}
