package cmd

import (
	"context"
	"fmt"
	"os"

	guardianv1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1"
	"github.com/odpf/guardian/app"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func appealsCommand(c *app.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "appeals",
		Short: "manage appeals",
	}

	cmd.AddCommand(listAppealsCommand(c))
	cmd.AddCommand(createAppealCommand(c))
	cmd.AddCommand(revokeAppealCommand(c))
	cmd.AddCommand(approveApprovalStepCommand(c))
	cmd.AddCommand(rejectApprovalStepCommand(c))

	return cmd
}

func listAppealsCommand(c *app.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list appeals",
		RunE: func(cmd *cobra.Command, args []string) error {
			cs := term.NewColorScheme()

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListAppeals(ctx, &guardianv1.ListAppealsRequest{})
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
}

func createAppealCommand(c *app.CLIConfig) *cobra.Command {
	var accountID string
	var resourceID uint
	var role string
	var optionsDuration string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create appeal",
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

			res, err := client.CreateAppeal(ctx, &guardianv1.CreateAppealRequest{
				AccountId: accountID,
				Resources: []*guardianv1.CreateAppealRequest_Resource{
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

	cmd.Flags().StringVar(&accountID, "account-id", "", "user email")
	cmd.MarkFlagRequired("account-id")
	cmd.Flags().UintVar(&resourceID, "resource-id", 0, "resource id")
	cmd.MarkFlagRequired("resource-id")
	cmd.Flags().StringVarP(&role, "role", "r", "", "role")
	cmd.MarkFlagRequired("role")
	cmd.Flags().StringVar(&optionsDuration, "options.duration", "", "access duration")

	return cmd
}

func revokeAppealCommand(c *app.CLIConfig) *cobra.Command {
	var id uint
	var reason string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "revoke an active access/appeal",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			_, err = client.RevokeAppeal(ctx, &guardianv1.RevokeAppealRequest{
				Id: uint32(id),
				Reason: &guardianv1.RevokeAppealRequest_Reason{
					Reason: reason,
				},
			})
			if err != nil {
				return err
			}

			fmt.Printf("appeal with id %v revoked successfully", id)

			return nil
		},
	}

	cmd.Flags().UintVar(&id, "id", 0, "appeal id")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVarP(&reason, "reason", "r", "", "rejection reason")

	return cmd
}

func approveApprovalStepCommand(c *app.CLIConfig) *cobra.Command {
	var id uint
	var approvalName string

	cmd := &cobra.Command{
		Use:   "approve",
		Short: "approve an approval step",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			_, err = client.UpdateApproval(ctx, &guardianv1.UpdateApprovalRequest{
				Id:           uint32(id),
				ApprovalName: approvalName,
				Action: &guardianv1.UpdateApprovalRequest_Action{
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

	cmd.Flags().UintVar(&id, "id", 0, "appeal id")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVarP(&approvalName, "approval-name", "a", "", "approval name going to be approved")
	cmd.MarkFlagRequired("approval-name")

	return cmd
}

func rejectApprovalStepCommand(c *app.CLIConfig) *cobra.Command {
	var id uint
	var approvalName string

	cmd := &cobra.Command{
		Use:   "reject",
		Short: "reject an approval step",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			_, err = client.UpdateApproval(ctx, &guardianv1.UpdateApprovalRequest{
				Id:           uint32(id),
				ApprovalName: approvalName,
				Action: &guardianv1.UpdateApprovalRequest_Action{
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

	cmd.Flags().UintVar(&id, "id", 0, "appeal id")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVarP(&approvalName, "approval-name", "a", "", "approval name going to be approved")
	cmd.MarkFlagRequired("approval-name")

	return cmd
}
