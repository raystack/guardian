package cli

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/go-playground/validator/v10"
	"github.com/raystack/guardian/internal/server"
	"github.com/raystack/guardian/jobs"
	"github.com/raystack/guardian/pkg/crypto"
	"github.com/raystack/guardian/plugins/notifiers"
	"github.com/raystack/salt/log"
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
			$ guardian job run expiring_grant_notification
			$ guardian job run revoke_expired_grants
			$ guardian job run revoke_grants_by_user_criteria
		`),
		Args: cobra.ExactValidArgs(1),
		ValidArgs: []string{
			string(jobs.TypeFetchResources),
			string(jobs.TypeExpiringGrantNotification),
			string(jobs.TypeRevokeExpiredGrants),
			string(jobs.TypeRevokeGrantsByUserCriteria),

			string(jobs.TypeRevokeExpiredAccess),
			string(jobs.TypeExpiringAccessNotification),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, err := cmd.Flags().GetString("config")
			if err != nil {
				return fmt.Errorf("getting config flag value: %w", err)
			}
			config, err := server.LoadConfig(configFile)
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

			services, err := server.InitServices(server.ServiceDeps{
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
				services.GrantService,
				services.ProviderService,
				notifier,
				crypto,
				validator,
			)

			jobsMap := map[jobs.Type]*struct {
				handler func(context.Context, jobs.Config) error
				config  jobs.Config
			}{
				jobs.TypeFetchResources: {
					handler: handler.FetchResources,
					config:  config.Jobs.FetchResources.Config,
				},
				jobs.TypeExpiringGrantNotification: {
					handler: handler.GrantExpirationReminder,
					config:  config.Jobs.ExpiringGrantNotification.Config,
				},
				jobs.TypeRevokeExpiredGrants: {
					handler: handler.RevokeExpiredGrants,
					config:  config.Jobs.RevokeExpiredGrants.Config,
				},
				jobs.TypeRevokeGrantsByUserCriteria: {
					handler: handler.RevokeGrantsByUserCriteria,
					config:  config.Jobs.RevokeGrantsByUserCriteria.Config,
				},

				// deprecated job names
				jobs.TypeExpiringAccessNotification: {
					handler: handler.GrantExpirationReminder,
					config:  config.Jobs.ExpiringAccessNotification.Config,
				},
				jobs.TypeRevokeExpiredAccess: {
					handler: handler.RevokeExpiredGrants,
					config:  config.Jobs.RevokeExpiredAccess.Config,
				},
			}

			jobName := jobs.Type(args[0])
			job := jobsMap[jobName]
			if job == nil {
				return fmt.Errorf("invalid job name: %s", jobName)
			}
			if err := job.handler(context.Background(), job.config); err != nil {
				return fmt.Errorf(`failed to run job "%s": %w`, jobName, err)
			}

			return nil
		},
	}

	return cmd
}
