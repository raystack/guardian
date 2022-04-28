package cmd

import (
	"fmt"

	"github.com/odpf/guardian/app"
	"github.com/odpf/salt/term"
	"github.com/odpf/salt/version"
	"github.com/spf13/cobra"
)

// VersionCmd prints the version of the binary
func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cs := term.NewColorScheme()

			if app.Version == "" {
				fmt.Println(cs.Yellow("guardian version (built from source)"))
				return nil
			}

			fmt.Printf("guardian version %s (%s)\n\n", app.Version, app.BuildDate)
			fmt.Println(cs.Yellow(version.UpdateNotice(app.Version, "odpf/guardian")))
			return nil
		},
	}
}
