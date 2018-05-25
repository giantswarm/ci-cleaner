package main

import (
	"log"

	"github.com/giantswarm/ci-cleaner/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
