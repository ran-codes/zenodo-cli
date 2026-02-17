package main

import (
	"os"

	"github.com/ran-codes/zenodo-cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
