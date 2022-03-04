package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/odpf/guardian/cmd"
	"github.com/odpf/salt/config"
)

const (
	exitOK    = 0
	exitError = 1
)

func main() {
	root := cmd.New()

	cliConfig, err := cmd.LoadConfig()
	if err != nil && !errors.As(err, &config.ConfigFileNotFoundError{}) {
		panic(err)
	}
	if cliConfig == nil {
		cliConfig = &cmd.Config{}
	}
	if err := cmd.BindFlagsFromConfig(root, cliConfig); err != nil {
		panic(err)
	}

	if cmd, err := root.ExecuteC(); err != nil {
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
}
