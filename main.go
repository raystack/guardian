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
		//fmt.Fprintln(os.Stderr, err)
		printError(err)

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

func printError(err error) {
	e := err.Error()
	if strings.Split(e, ":")[0] == "rpc error" {
		es := strings.Split(e, "= ")

		em := es[2]
		errMsg := "Error: " + em

		ec := es[1]
		errCode := ec[0 : len(ec)-5]
		errCode = "Code: " + errCode

		fmt.Fprintln(os.Stderr, errCode, errMsg)
	} else {
		fmt.Fprintln(os.Stderr, err)
	}

	return
}
