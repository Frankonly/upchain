package main

import (
	"log"

	"upchain/cli"
)

func main() {
	if err := cli.Init(); err != nil {
		log.Fatalf("failed to initialize upcli: %v", err)
	}

	cli.Execute()
}
