package main

import (
	"os"

	"github.com/rikurb8/carnie/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
