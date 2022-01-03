package main

import (
	"fmt"
	"os"

	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/cmd"
	"github.com/odpf/salt/term"
)

const (
	exitOK    = 0
	exitError = 1
)

func main() {
	cliConfig, err := app.LoadCLIConfig()
	if err != nil {
		// Init with blank client config
		cliConfig = &app.CLIConfig{}
		cs := term.NewColorScheme()

		defer fmt.Println(cs.Yellow("client not configured. try running `guardian config init`"))
	}

	command := cmd.New(cliConfig)

	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitError)
	}
}
