package main

import (
	"fmt"
	"os"

	"github.com/odpf/guardian/cmd"
)

const (
	exitOK    = 0
	exitError = 1
)

func main() {

	command := cmd.New()

	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitError)
	}

}
