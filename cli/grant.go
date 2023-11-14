package cli

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	handlerv1beta1 "github.com/raystack/guardian/api/handler/v1beta1"
	guardianv1beta1 "github.com/raystack/guardian/api/proto/raystack/guardian/v1beta1"
	"github.com/raystack/salt/printer"
	"github.com/raystack/salt/term"
	"github.com/spf13/cobra"
)

func grantsCommand(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "grant",
		Aliases: []string{"grants"},
		Short:   "Manage grants",
		Annotations: map[string]string{
			"group": "core",
		},
		Example: heredoc.Doc(`
			$ guardian grant list
			$ guardian grant view
			$ guardian grant revoke
			$ guardian grant bulk-revoke
		`),
	}

	cmd.AddCommand(listGrantsCommand())
	cmd.AddCommand(viewGrantCommand(adapter))
	cmd.AddCommand(revokeGrantCommand())
	cmd.AddCommand(bulkRevokeGrantCommand())
	bindFlagsFromConfig(cmd)

	return cmd
}

func listGrantsCommand() *cobra.Command {
	var statuses, accountIDs, accountTypes, resourceIDs, roles, permissions,
		providerTypes, providerURNs, resourceTypes, resourceURNs []string
	var createdBy string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List and filter grants",
		Example: heredoc.Doc(`
			$ guardian grant list
			$ guardian grant list --statuses=pending
			$ guardian grant list --roles=viewer
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListGrants(createCtx(cmd), &guardianv1beta1.ListGrantsRequest{
				Statuses:      statuses,
				AccountIds:    accountIDs,
				AccountTypes:  accountTypes,
				ResourceIds:   resourceIDs,
				ProviderTypes: providerTypes,
				ProviderUrns:  providerURNs,
				ResourceTypes: resourceTypes,
				ResourceUrns:  resourceURNs,
				Roles:         roles,
				CreatedBy:     createdBy,
			})
			if err != nil {
				return err
			}
			spinner.Stop()

			report := [][]string{}

			grants := res.GetGrants()
			report = append(report, []string{"ID", "ACCOUNT ID", "RESOURCE ID", "ROLE", "STATUS"})
			for _, g := range grants {
				report = append(report, []string{
					term.Greenf("%v", g.GetId()),
					g.GetAccountId(),
					g.GetResourceId(),
					g.GetRole(),
					g.GetStatus(),
				})
			}
			printer.Table(os.Stdout, report)
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&statuses, "statuses", nil, "Filter by grant status")
	cmd.Flags().StringArrayVar(&accountIDs, "account-ids", nil, "Filter by account id")
	cmd.Flags().StringArrayVar(&accountTypes, "account-types", nil, "Filter by account type")
	cmd.Flags().StringArrayVar(&resourceIDs, "resource-ids", nil, "Filter by resource id")
	cmd.Flags().StringArrayVar(&roles, "roles", nil, "Filter by role")
	cmd.Flags().StringArrayVar(&permissions, "permissions", nil, "Filter by permissions")
	cmd.Flags().StringArrayVar(&providerTypes, "provider-types", nil, "Filter by resource provider type")
	cmd.Flags().StringArrayVar(&providerURNs, "provider-urns", nil, "Filter by resource provider urn")
	cmd.Flags().StringArrayVar(&resourceTypes, "resource-types", nil, "Filter by resource type")
	cmd.Flags().StringArrayVar(&resourceURNs, "resource-urns", nil, "Filter by resource urn")
	cmd.Flags().StringVar(&createdBy, "created-by", "", "Filter by creator")

	return cmd
}

func viewGrantCommand(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Get grant details",
		Example: heredoc.Doc(`
			$ guardian grant view <grant-id>
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			id := args[0]
			res, err := client.GetGrant(createCtx(cmd), &guardianv1beta1.GetGrantRequest{
				Id: id,
			})
			if err != nil {
				return err
			}
			spinner.Stop()

			g := adapter.FromGrantProto(res.GetGrant())
			if err := printer.File(g, format); err != nil {
				return fmt.Errorf("failed to format grant result: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "output", "o", "yaml", "Print output with the selected format")

	return cmd
}

func revokeGrantCommand() *cobra.Command {
	var reason string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke an active grant",
		Example: heredoc.Doc(`
			$ guardian grant revoke <grant-id>
			$ guardian grant revoke <grant-id> --reason=<reason>
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			id := args[0]
			if _, err := client.RevokeGrant(createCtx(cmd), &guardianv1beta1.RevokeGrantRequest{
				Id:     id,
				Reason: reason,
			}); err != nil {
				return err
			}

			spinner.Stop()

			fmt.Printf("grant with id %q revoked successfully", id)
			return nil
		},
	}

	cmd.Flags().StringVarP(&reason, "reason", "r", "", "Reason of the revocation")
	return cmd
}

func bulkRevokeGrantCommand() *cobra.Command {
	var accountIds []string
	var providerTypes []string
	var providerUrns []string
	var resourceTypes []string
	var resourceUrns []string
	var reason string

	cmd := &cobra.Command{
		Use:   "bulk-revoke",
		Short: "Bulk Revoke active grants",
		Example: heredoc.Doc(`
			$ guardian grant bulk-revoke
			$ guardian grant bulk-revoke --account-ids=<account-ids> --reason=<reason> --provider-types=<provider-types> --provider-urns=<provider-urns>
			$ guardian grant bulk-revoke --account-ids=<account-ids> --reason=<reason> --resource-types=<resource-types> --resource-urns=<resource-urns>
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			response, err := client.RevokeGrants(createCtx(cmd), &guardianv1beta1.RevokeGrantsRequest{
				AccountIds:    accountIds,
				ProviderTypes: providerTypes,
				ProviderUrns:  providerUrns,
				ResourceTypes: resourceTypes,
				ResourceUrns:  resourceUrns,
				Reason:        reason,
			})
			if err != nil {
				return err
			}

			var report [][]string
			grants := response.GetGrants()
			spinner.Stop()

			fmt.Printf(" \nShowing %d revoked grants of account-ids: %v \n \n", len(grants), len(accountIds))

			report = append(report, []string{"ID", "USER", "RESOURCE ID", "ROLE", "STATUS"})
			for _, g := range grants {
				report = append(report, []string{
					term.Greenf("%v", g.GetId()),
					g.GetAccountId(),
					fmt.Sprintf("%v", g.GetResourceId()),
					g.GetRole(),
					g.GetStatus(),
				})
			}
			printer.Table(os.Stdout, report)
			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&accountIds, "account-ids", "a", nil, "Filter by accountIds")
	cmd.Flags().StringArrayVar(&providerTypes, "provider-types", nil, "Filter by providerTypes")
	cmd.Flags().StringArrayVar(&providerUrns, "provider-urns", nil, "Filter by providerUrns")
	cmd.Flags().StringArrayVar(&resourceTypes, "resource-types", nil, "Filter by resourceTypes")
	cmd.Flags().StringArrayVar(&resourceUrns, "resource-urns", nil, "Filter by resourceUrns")
	cmd.Flags().StringVarP(&reason, "reason", "r", "", "Reason of the revocation")
	cmd.MarkFlagRequired("account-ids")
	cmd.MarkFlagRequired("reason")
	return cmd
}
