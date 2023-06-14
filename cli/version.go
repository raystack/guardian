package cli

import (
	"fmt"

	"github.com/raystack/guardian/core"
	"github.com/raystack/salt/term"
	"github.com/raystack/salt/version"
	"github.com/spf13/cobra"
)

// VersionCmd prints the version of the binary
func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if core.Version == "" {
				fmt.Println(term.Yellow("guardian version (built from source)"))
				return nil
			}

			fmt.Printf("guardian version %s (%s)\n\n", core.Version, core.BuildDate)
			fmt.Println(term.Yellow(version.UpdateNotice(core.Version, "raystack/guardian")))
			return nil
		},
	}
}
