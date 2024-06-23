//go:build debug

package main

import (
	"github.com/memocash/index/cmd"
	"github.com/memocash/index/cmd/test"
	"log"
)

func main() {
	if err := cmd.Execute(
		test.GetCommand(),
	); err != nil {
		log.Fatalf("fatal error executing debug command; %v", err)
	}
}
