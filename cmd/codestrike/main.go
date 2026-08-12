package main

import (
	"os"

	"github.com/CrowdStrike/codestrike/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
