package main

import (
	"fmt"
	"log"
	"syne-cli/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(fmt.Sprintf("%s (commit %s built at %s)", version, commit, date))

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
