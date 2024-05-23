package main

import (
	"github.com/memocash/index/cmd"
	"log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("fatal error executing command; %v", err)
	}
}
