package cli

import (
	"github.com/MakeNowJust/heredoc"
)

var envHelp = map[string]string{
	"short": "List of supported environment variables",
	"long": heredoc.Doc(`
			GOTO_CONFIG_DIR: the directory where guardian will store configuration files. Default:
			"$XDG_CONFIG_HOME/goto" or "$HOME/.config/goto".

			NO_COLOR: set to any value to avoid printing ANSI escape sequences for color output.

			CLICOLOR: set to "0" to disable printing ANSI colors in output.
		`),
}
