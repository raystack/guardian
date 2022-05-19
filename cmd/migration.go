package cmd

import (
	"fmt"
	"github.com/MakeNowJust/heredoc"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/crypto"
	"github.com/odpf/guardian/plugins/migrations/metabase"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func MigrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migration",
		Aliases: []string{"migration"},
		Short:   "Manage migration",
		Long: heredoc.Doc(`
			Work with migrations.

			Migrations are used to populate guardian appeals from target system`),
		Example: heredoc.Doc(`
			$ guardian migration metabase
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
	}

	cmd.AddCommand(metabaseMigrationCmd())
	bindFlagsFromConfig(cmd)

	return cmd
}

func metabaseMigrationCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "metabase",
		Short: "Metabase migration",
		Long: heredoc.Doc(`
			List and filter all available access policies.
		`),
		Example: heredoc.Doc(`
			$ guardian migration metabase <provider-urn> <policy-id>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			providerId := args[0]
			policyId := args[1]
			fmt.Println(providerId, policyId)

			say := crypto.NewAES("")
			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			context := cmd.Context()
			providerResponse, err := client.GetProvider(context, &guardianv1beta1.GetProviderRequest{Id: providerId})

			provider := providerResponse.GetProvider()
			providerUrn := provider.Urn
			fields := provider.Config.Credentials.GetStructValue().GetFields()
			abc, err := say.Decrypt(fields["password"].GetStringValue())

			providerConfig := domain.ProviderConfig{
				Type: provider.Config.Type,
				URN:  provider.Config.Urn,
				Credentials: map[string]string{
					"username": fields["username"].GetStringValue(),
					"password": abc,
					"host":     fields["host"].GetStringValue(),
				},
				Appeal:    nil,
				Resources: nil,
			}

			listResources, err := client.ListResources(context, &guardianv1beta1.ListResourcesRequest{ProviderUrn: providerUrn, IsDeleted: false})
			resources := make([]domain.Resource, 0)
			for _, r := range listResources.GetResources() {
				resources = append(resources, domain.Resource{
					ID:           r.Id,
					ProviderType: r.ProviderType,
					ProviderURN:  r.ProviderUrn,
					Type:         r.Type,
					URN:          r.Urn,
					Name:         r.Name,
				})
			}

			listAppeals, err := client.ListAppeals(context, &guardianv1beta1.ListAppealsRequest{ProviderUrns: []string{providerUrn}, Statuses: []string{"pending"}})

			appeals := make([]domain.Appeal, 0)
			for _, a := range listAppeals.GetAppeals() {
				appeals = append(appeals, domain.Appeal{
					ID:          a.Id,
					ResourceID:  a.ResourceId,
					Status:      a.Status,
					AccountID:   a.AccountId,
					AccountType: a.AccountType,
					Role:        a.Role,
				})
			}

			migration := metabase.NewMigration(providerConfig, resources, appeals)
			appealRequests, err := migration.PopulateAccess()

			for _, appealRequest := range appealRequests {
				resource := appealRequest.Resource
				option, _ := structpb.NewStruct(map[string]interface{}{"duration": resource.Duration})

				appeal, _ := client.CreateAppeal(context, &guardianv1beta1.CreateAppealRequest{
					AccountId: appealRequest.AccountID,
					Resources: []*guardianv1beta1.CreateAppealRequest_Resource{{Id: resource.ID,
						Options: option}},
					AccountType: "",
				})

				fmt.Println(appeal)
				//Todo: iterate appeal's approval and approve it via #{client.UpdateApproval}
			}

			return nil
		},
	}
}
