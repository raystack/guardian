package cli

import (
	"fmt"

	"github.com/odpf/salt/term"
	"github.com/odpf/salt/version"
	"github.com/spf13/cobra"
)

var (
	// Version is the version of the binary
	Version string
	// BuildCommit is the commit hash of the binary
	BuildCommit string
	// BuildDate is the date of the build
	BuildDate string
)

// VersionCmd prints the version of the binary
func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cs := term.NewColorScheme()

			if Version == "" {
				fmt.Println(cs.Yellow("guardian version (built from source)"))
				return nil
			}

			fmt.Printf("guardian version %s (%s)\n\n", Version, BuildDate)
			fmt.Println(cs.Yellow(version.UpdateNotice(Version, "odpf/guardian")))
			return nil
		},
	}
}
