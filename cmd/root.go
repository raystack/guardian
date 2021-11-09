package cmd

import (
	"github.com/MakeNowJust/heredoc"
	handlerv1beta1 "github.com/odpf/guardian/api/handler/v1beta1"
	"github.com/odpf/guardian/app"
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"
)

//New  create a root command
func New() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "guardian <command> <subcommand> [flags]",
		Short: "Universal data access control",
		Long: heredoc.Doc(`
			Universal data access control.

			Guardian is a tool for extensible and universal data access with 
			automated access workflows and security controls across data stores, 
			analytical systems, and cloud products.`),
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: heredoc.Doc(`
			$ guardian policies list
			$ guardian providers list
			$ guardian resources list
			$ guardian policies create --file policy.yaml
		`),
		Annotations: map[string]string{
			"group:core": "true",
			"help:learn": heredoc.Doc(`
				Use 'guardian <command> <subcommand> --help' for more information about a command.
				Read the manual at https://odpf.gitbook.io/guardian/
			`),
			"help:feedback": heredoc.Doc(`
				Open an issue here https://github.com/odpf/guardian/issues
			`),
		},
	}

	protoAdapter := handlerv1beta1.NewAdapter()

	cliConfig, err := app.LoadCLIConfig(app.CLIConfigFile)
	if err != nil {
		panic(err)
	}

	cmdx.SetHelp(cmd)

	cmd.AddCommand(serveCommand())
	cmd.AddCommand(migrateCommand())
	cmd.AddCommand(configCommand())
	cmd.AddCommand(ResourceCmd(cliConfig, protoAdapter))
	cmd.AddCommand(ProviderCmd(cliConfig, protoAdapter))
	cmd.AddCommand(PolicyCmd(cliConfig, protoAdapter))
	cmd.AddCommand(appealsCommand(cliConfig))

	return cmd
}
