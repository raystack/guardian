package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	. "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/crypto"
	"github.com/odpf/guardian/plugins/migrations"
	mb "github.com/odpf/guardian/plugins/migrations/metabase"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	pending = "pending"
	active  = "active"
)

func MigrationCmd(config *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Guardian migration",
		Long: heredoc.Doc(`
			Migrate target system ACL into Guardian.
		`),
		Example: heredoc.Doc(`
			$ guardian migration <provider-urn>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			providerId := args[0]

			say := crypto.NewAES(config.EncryptionSecretKey)
			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			context := cmd.Context()
			provider, err := getProviderConfig(client, context, providerId, say)
			if err != nil {
				return err
			}

			resources, err := getResources(client, context, provider)
			if err != nil {
				return err
			}

			appealResponse, appeals, err := getActiveAndPendingAppeals(client, context, provider)
			if err != nil {
				return err
			}

			var migration migrations.Client
			if provider.Type == migrations.Metabase {
				migration = mb.NewMigration(provider.Config, resources, appeals)
			} else {
				return errors.New(fmt.Sprintf("Migration not supported for provider %s", provider.Type))
			}

			appealRequests, err := migration.PopulateAccess()
			if err != nil {
				return err
			}

			//migrate past-run pending appeals
			for _, a := range appealResponse {
				if a.Status == pending {
					err := approveAppeal(a, client, context)
					if err != nil {
						return err
					} else {
						fmt.Println(a.Resource.Name, a.AccountId)
					}
				}
			}

			//migrate pending appeals
			for _, appealRequest := range appealRequests {
				resource := appealRequest.Resource
				option, _ := structpb.NewStruct(map[string]interface{}{migrations.Duration: resource.Duration})

				accountID := appealRequest.AccountID
				appeal, err := client.CreateAppeal(context, &CreateAppealRequest{
					AccountId: accountID,
					Resources: []*CreateAppealRequest_Resource{
						{Id: resource.ID, Role: resource.Role, Options: option}},
					AccountType: "",
				})
				if err != nil {
					return err
				} else {
					appeals := appeal.GetAppeals()
					for _, appeal := range appeals {
						err := approveAppeal(appeal, client, context)
						if err != nil {
							return err
						}
					}
				}
			}

			return nil
		},
	}

	bindFlagsFromConfig(cmd)

	return cmd
}

func getActiveAndPendingAppeals(client GuardianServiceClient, context context.Context, provider *domain.Provider) ([]*Appeal, []domain.Appeal, error) {
	appeals := make([]domain.Appeal, 0)
	listAppeals, err := client.ListAppeals(context, &ListAppealsRequest{ProviderUrns: []string{provider.URN}, Statuses: []string{pending, active}})
	if err != nil {
		return nil, appeals, err
	}

	appealResponses := listAppeals.GetAppeals()
	for _, a := range appealResponses {
		appeals = append(appeals, domain.Appeal{
			ID:          a.Id,
			ResourceID:  a.ResourceId,
			Status:      a.Status,
			AccountID:   a.AccountId,
			AccountType: a.AccountType,
			Role:        a.Role,
		})
	}
	return appealResponses, appeals, nil
}

func getResources(client GuardianServiceClient, context context.Context, provider *domain.Provider) ([]domain.Resource, error) {
	listResources, err := client.ListResources(context, &ListResourcesRequest{ProviderUrn: provider.URN, IsDeleted: false})
	resources := make([]domain.Resource, 0)
	if err != nil {
		return resources, err
	}
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
	return resources, nil
}

func getProviderConfig(client GuardianServiceClient, context context.Context, providerId string, say *crypto.AES) (*domain.Provider, error) {
	providerResponse, err := client.GetProvider(context, &GetProviderRequest{Id: providerId})
	if err != nil {
		return nil, err
	}

	provider := providerResponse.GetProvider()
	fields := provider.Config.Credentials.GetStructValue().GetFields()
	abc, err := say.Decrypt(fields[migrations.Password].GetStringValue())

	return &domain.Provider{
		ID:   providerId,
		Type: provider.Type,
		URN:  provider.Urn,
		Config: &domain.ProviderConfig{
			Type: provider.Config.Type,
			URN:  provider.Config.Urn,
			Credentials: map[string]string{
				migrations.Username: fields[migrations.Username].GetStringValue(),
				migrations.Password: abc,
				migrations.Host:     fields[migrations.Host].GetStringValue(),
			},
		},
	}, err
}

func approveAppeal(appeal *Appeal, client GuardianServiceClient, context context.Context) error {
	approvals := appeal.Approvals
	for _, approval := range approvals {
		if approval.Status == pending {
			_, err := client.UpdateApproval(context, &UpdateApprovalRequest{
				Id:           appeal.Id,
				ApprovalName: approval.Name,
				Action:       &UpdateApprovalRequest_Action{Action: "approve", Reason: "Metabase migration"},
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
