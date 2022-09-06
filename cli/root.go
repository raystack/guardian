package cli

import (
	"github.com/MakeNowJust/heredoc"
	handlerv1beta1 "github.com/odpf/guardian/api/handler/v1beta1"
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"
)

func New(cfg *Config) *cobra.Command {
	cliConfig = cfg
	var cmd = &cobra.Command{
		Use:   "guardian <command> <subcommand> [flags]",
		Short: "Universal data access control",
		Long: heredoc.Doc(`
			Universal access control to cloud apps and infrastructure.`),
		SilenceUsage:  true,
		SilenceErrors: true,
		Annotations: map[string]string{
			"group": "core",
			"help:learn": heredoc.Doc(`
				Use 'guardian <command> --help' for info about a command.
				Read the manual at https://odpf.github.io/guardian/
			`),
			"help:feedback": heredoc.Doc(`
				Open an issue here https://github.com/odpf/guardian/issues
			`),
		},
	}

	protoAdapter := handlerv1beta1.NewAdapter()
	cmd.AddCommand(ResourceCmd(protoAdapter))
	cmd.AddCommand(ProviderCmd(protoAdapter))
	cmd.AddCommand(PolicyCmd(protoAdapter))
	cmd.AddCommand(appealsCommand())
	cmd.AddCommand(grantsCommand(protoAdapter))
	cmd.AddCommand(ServerCommand())
	cmd.AddCommand(JobCmd())
	cmd.AddCommand(configCommand())
	cmd.AddCommand(VersionCmd())

	// Help topics
	cmdx.SetHelp(cmd)
	cmd.AddCommand(cmdx.SetCompletionCmd("guardian"))
	cmd.AddCommand(cmdx.SetHelpTopicCmd("environment", envHelp))
	cmd.AddCommand(cmdx.SetRefCmd(cmd))

	return cmd
}
