package main

import (
	"fmt"
	"os"
	"strings"

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

	root := cmd.New(cliConfig)

	if cmd, err := root.ExecuteC(); err != nil {
		fmt.Fprintln(os.Stderr, err)

		cmdErr := strings.HasPrefix(err.Error(), "unknown command")
		flagErr := strings.HasPrefix(err.Error(), "unknown flag")
		sflagErr := strings.HasPrefix(err.Error(), "unknown shorthand flag")

		if cmdErr || flagErr || sflagErr {
			if !strings.HasSuffix(err.Error(), "\n") {
				fmt.Println()
			}
			fmt.Println(cmd.UsageString())
			os.Exit(exitOK)
		} else {
			os.Exit(exitError)
		}
	}

	// TODO: Notify if updated version is available
}
