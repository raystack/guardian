package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/app"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func appealsCommand(c *app.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "appeal",
		Aliases: []string{"appeals"},
		Short:   "Manage appeals",
		Annotations: map[string]string{
			"group:core": "true",
		},
		Example: heredoc.Doc(`
			$ guardian appeal create
			$ guardian appeal approve
			$ guardian appeal list --status=pending
		`),
	}

	cmd.AddCommand(listAppealsCommand(c))
	cmd.AddCommand(createAppealCommand(c))
	cmd.AddCommand(revokeAppealCommand(c))
	cmd.AddCommand(approveApprovalStepCommand(c))
	cmd.AddCommand(rejectApprovalStepCommand(c))

	return cmd
}

func listAppealsCommand(c *app.CLIConfig) *cobra.Command {
	var statuses []string
	var role string
	var accountID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List and filter appeals",
		Example: heredoc.Doc(`
			$ guardian appeal list
			$ guardian appeal list --status=pending
			$ guardian appeal list --role=viewer
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cs := term.NewColorScheme()

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListAppeals(ctx, &guardianv1beta1.ListAppealsRequest{
				Statuses:  statuses,
				Role:      role,
				AccountId: accountID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}

			appeals := res.GetAppeals()
			fmt.Printf(" \nShowing %d of %d policies\n \n", len(appeals), len(appeals))

			report = append(report, []string{"ID", "USER", "RESOURCE ID", "ROLE", "STATUS"})
			for _, a := range appeals {
				report = append(report, []string{
					cs.Greenf("%v", a.GetId()),
					a.GetAccountId(),
					fmt.Sprintf("%v", a.GetResourceId()),
					a.GetRole(),
					a.GetStatus(),
				})
			}
			printer.Table(os.Stdout, report)
			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&statuses, "status", "s", nil, "Filter by status(es)")
	cmd.Flags().StringVarP(&role, "role", "r", "", "Filter by role")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Filter by account")

	return cmd
}

func createAppealCommand(c *app.CLIConfig) *cobra.Command {
	var accountID, accountType string
	var resourceID uint
	var role string
	var optionsDuration string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new appeal",
		Example: heredoc.Doc(`
			$ guardian appeal create
			$ guardian appeal create --account=<account-id> --type=<account-type> --resource=<resource-id> --role=<role>
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := map[string]interface{}{}
			if optionsDuration != "" {
				options["duration"] = optionsDuration
			}
			optionsProto, err := structpb.NewStruct(options)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.CreateAppeal(ctx, &guardianv1beta1.CreateAppealRequest{
				AccountId:   accountID,
				AccountType: accountType,
				Resources: []*guardianv1beta1.CreateAppealRequest_Resource{
					{
						Id:      uint32(resourceID),
						Role:    role,
						Options: optionsProto,
					},
				},
			})
			if err != nil {
				return err
			}

			appealID := res.GetAppeals()[0].GetId()
			fmt.Printf("appeal created with id: %v", appealID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Email of the account to appeal")
	cmd.MarkFlagRequired("account")

	cmd.Flags().StringVarP(&accountType, "type", "t", "", "Type of the account")
	cmd.MarkFlagRequired("type")

	cmd.Flags().UintVarP(&resourceID, "resource", "R", 0, "ID of the resource")
	cmd.MarkFlagRequired("resource")

	cmd.Flags().StringVarP(&role, "role", "r", "", "Role to be assigned")
	cmd.MarkFlagRequired("role")

	cmd.Flags().StringVarP(&optionsDuration, "duration", "d", "", "Duration of the access")

	return cmd
}

func revokeAppealCommand(c *app.CLIConfig) *cobra.Command {
	var id uint
	var reason string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke an active access/appeal",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			_, err = client.RevokeAppeal(ctx, &guardianv1beta1.RevokeAppealRequest{
				Id: uint32(id),
				Reason: &guardianv1beta1.RevokeAppealRequest_Reason{
					Reason: reason,
				},
			})
			if err != nil {
				return err
			}

			fmt.Printf("appeal with id `%v` revoked successfully", id)

			return nil
		},
	}

	cmd.Flags().UintVar(&id, "id", 0, "ID of the appeal")
	cmd.MarkFlagRequired("id")

	cmd.Flags().StringVarP(&reason, "reason", "r", "", "Reason of the revocation")

	return cmd
}

func approveApprovalStepCommand(c *app.CLIConfig) *cobra.Command {
	var id uint
	var approvalName string

	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve an approval step",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			_, err = client.UpdateApproval(ctx, &guardianv1beta1.UpdateApprovalRequest{
				Id:           uint32(id),
				ApprovalName: approvalName,
				Action: &guardianv1beta1.UpdateApprovalRequest_Action{
					Action: "approve",
				},
			})
			if err != nil {
				return err
			}

			fmt.Printf("appeal with id %v and approval name %v approved successfully", id, approvalName)

			return nil
		},
	}

	cmd.Flags().UintVar(&id, "id", 0, "ID of the appeal")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVarP(&approvalName, "step", "s", "", "Name of approval step")
	cmd.MarkFlagRequired("approval-name")

	return cmd
}

func rejectApprovalStepCommand(c *app.CLIConfig) *cobra.Command {
	var id uint
	var approvalName string

	cmd := &cobra.Command{
		Use:   "reject",
		Short: "Reject an approval step",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			_, err = client.UpdateApproval(ctx, &guardianv1beta1.UpdateApprovalRequest{
				Id:           uint32(id),
				ApprovalName: approvalName,
				Action: &guardianv1beta1.UpdateApprovalRequest_Action{
					Action: "reject",
				},
			})
			if err != nil {
				return err
			}

			fmt.Printf("appeal with id %v and approval name %v rejected successfully", id, approvalName)

			return nil
		},
	}

	cmd.Flags().UintVar(&id, "id", 0, "ID of the appeal")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVarP(&approvalName, "step", "s", "", "Name of approval step")
	cmd.MarkFlagRequired("approval-name")

	return cmd
}
