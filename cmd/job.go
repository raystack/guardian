package cmd

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/internal/crypto"
	"github.com/odpf/guardian/jobs"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/salt/log"
	"github.com/spf13/cobra"
)

func JobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "job",
		Aliases: []string{"jobs"},
		Short:   "Manage jobs",
		Example: heredoc.Doc(`
			$ guardian job run fetch_resources
		`),
	}

	cmd.AddCommand(
		runJobCmd(),
	)

	cmd.PersistentFlags().StringP("config", "c", "./config.yaml", "Config file path")
	cmd.MarkPersistentFlagFilename("config")

	return cmd
}

func runJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Fire a specific job",
		Example: heredoc.Doc(`
			$ guardian job run fetch_resources
			$ guardian job run appeal_expiration_reminder
			$ guardian job run appeal_expiration_revocation
		`),
		Args: cobra.ExactValidArgs(1),
		ValidArgs: []string{
			"fetch_resources",
			"appeal_expiration_reminder",
			"appeal_expiration_revocation",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, err := cmd.Flags().GetString("config")
			if err != nil {
				return fmt.Errorf("getting config flag value: %w", err)
			}
			config, err := app.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			logger := log.NewLogrus(log.LogrusWithLevel(config.LogLevel))
			crypto := crypto.NewAES(config.EncryptionSecretKeyKey)
			validator := validator.New()
			notifier, err := notifiers.NewClient(&config.Notifier)
			if err != nil {
				return err
			}

			services, err := app.InitServices(app.ServiceDeps{
				Config:    &config,
				Logger:    logger,
				Validator: validator,
				Notifier:  notifier,
				Crypto:    crypto,
			})
			if err != nil {
				return fmt.Errorf("initializing services: %w", err)
			}

			handler := jobs.NewHandler(
				logger,
				services.AppealService,
				services.ProviderService,
				notifier,
			)

			jobs := map[string]func(context.Context) error{
				"fetch_resources":              handler.FetchResources,
				"appeal_expiration_reminder":   handler.AppealExpirationReminder,
				"appeal_expiration_revocation": handler.RevokeExpiredAppeals,
			}

			jobName := args[0]
			job := jobs[jobName]
			if job == nil {
				return fmt.Errorf("invalid job name: %s", jobName)
			}
			if err := job(context.Background()); err != nil {
				return fmt.Errorf(`failed to run job "%s": %w`, jobName, err)
			}

			return nil
		},
	}

	return cmd
}
