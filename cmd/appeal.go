package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/domain"
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
			$ guardian appeal reject
			$ guardian appeal list
			$ guardian appeal status
			$ guardian appeal revoke
			$ guardian appeal cancel
		`),
	}

	cmd.AddCommand(listAppealsCommand(c))
	cmd.AddCommand(createAppealCommand(c))
	cmd.AddCommand(revokeAppealCommand(c))
	cmd.AddCommand(approveApprovalStepCommand(c))
	cmd.AddCommand(rejectApprovalStepCommand(c))
	cmd.AddCommand(statusAppealCommand(c))
	cmd.AddCommand(cancelAppealCommand(c))

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
			s := term.Spin("Fetching appeal list")
			defer s.Stop()

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
			s.Stop()

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
	var accountID, accountType, resourceID, role, optionsDuration string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new appeal",
		Example: heredoc.Doc(`
			$ guardian appeal create
			$ guardian appeal create --account=<account-id> --type=<account-type> --resource=<resource-id> --role=<role>
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := term.Spin("Creating appeal")
			defer s.Stop()

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
						Id:      resourceID,
						Role:    role,
						Options: optionsProto,
					},
				},
			})
			if err != nil {
				fmt.Println("Create error")
				return err
			}

			s.Stop()

			appealID := res.GetAppeals()[0].GetId()
			fmt.Printf("appeal created with id: %v", appealID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Email of the account to appeal")
	cmd.MarkFlagRequired("account")

	cmd.Flags().StringVarP(&accountType, "type", "t", "", "Type of the account")
	cmd.MarkFlagRequired("type")

	cmd.Flags().StringVarP(&resourceID, "resource", "R", "", "ID of the resource")
	cmd.MarkFlagRequired("resource")

	cmd.Flags().StringVarP(&role, "role", "r", "", "Role to be assigned")
	cmd.MarkFlagRequired("role")

	cmd.Flags().StringVarP(&optionsDuration, "duration", "d", "", "Duration of the access")

	return cmd
}

func revokeAppealCommand(c *app.CLIConfig) *cobra.Command {
	var reason string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke an active access/appeal",
		Example: heredoc.Doc(`
		$ guardian appeal revoke <appeal-id>
		$ guardian appeal revoke <appeal-id> --reason=<reason>
	`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := term.Spin("Revoking appeal")
			defer s.Stop()

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			id := args[0]
			_, err = strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid provider id: %v", err)
			}

			_, err = client.RevokeAppeal(ctx, &guardianv1beta1.RevokeAppealRequest{
				Id: id,
				Reason: &guardianv1beta1.RevokeAppealRequest_Reason{
					Reason: reason,
				},
			})
			if err != nil {
				return err
			}

			s.Stop()

			fmt.Printf("appeal with id `%v` revoked successfully", id)

			return nil
		},
	}

	cmd.Flags().StringVarP(&reason, "reason", "r", "", "Reason of the revocation")

	return cmd
}

func approveApprovalStepCommand(c *app.CLIConfig) *cobra.Command {
	var approvalName string

	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve an approval step",
		Example: heredoc.Doc(`
		$ guardian appeal approve <appeal-id> --step=<step-name>
	`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := term.Spin("Approving appeal")
			defer s.Stop()

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			id := args[0]
			_, err = strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid provider id: %v", err)
			}

			_, err = client.UpdateApproval(ctx, &guardianv1beta1.UpdateApprovalRequest{
				Id:           id,
				ApprovalName: approvalName,
				Action: &guardianv1beta1.UpdateApprovalRequest_Action{
					Action: "approve",
				},
			})
			if err != nil {
				return err
			}

			s.Stop()

			fmt.Printf("appeal with id %v and approval name %v approved successfully", id, approvalName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&approvalName, "step", "s", "", "Name of approval step")
	cmd.MarkFlagRequired("approval-name")

	return cmd
}

func rejectApprovalStepCommand(c *app.CLIConfig) *cobra.Command {
	var approvalName string

	cmd := &cobra.Command{
		Use:   "reject",
		Short: "Reject an approval step",
		Example: heredoc.Doc(`
		$ guardian appeal reject <appeal-id> --step=<step-name>
	`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := term.Spin("Rejecting appeal")
			defer s.Stop()

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			id := args[0]
			_, err = strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid provider id: %v", err)
			}

			_, err = client.UpdateApproval(ctx, &guardianv1beta1.UpdateApprovalRequest{
				Id:           id,
				ApprovalName: approvalName,
				Action: &guardianv1beta1.UpdateApprovalRequest_Action{
					Action: "reject",
				},
			})
			if err != nil {
				return err
			}

			s.Stop()

			fmt.Printf("appeal with id %v and approval name %v rejected successfully", id, approvalName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&approvalName, "step", "s", "", "Name of approval step")
	cmd.MarkFlagRequired("approval-name")

	return cmd
}

func statusAppealCommand(c *app.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Approval status of an appeal",
		Example: heredoc.Doc(`
			$ guardian appeal status <appeal-id>
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			id := args[0]
			_, err = strconv.ParseUint(id, 10, 32)
			if err != nil {
				return fmt.Errorf("invalid provider id: %v", err)
			}

			res, err := client.GetAppeal(ctx, &guardianv1beta1.GetAppealRequest{
				Id: id,
			})
			if err != nil {
				return err
			}

			appeal := res.GetAppeal()
			fmt.Printf(" \nAppeal status: %s\n", appeal.GetStatus())

			approvals := appeal.Approvals

			report := [][]string{}
			report = append(report, []string{"ID", "NAME", "STATUS", "APPROVER(s)"})

			fmt.Printf(" \nShowing %d approval steps\n \n", len(approvals))

			for _, a := range approvals {
				status := a.GetStatus()
				actor := a.GetActor()

				if actor != "" {
					if status == domain.ApprovalStatusApproved {
						status = fmt.Sprintf("approved by %s", actor)
					} else if status == domain.ApprovalStatusRejected {
						status = fmt.Sprintf("rejected by %s", actor)
					}
				}

				report = append(report, []string{
					fmt.Sprintf("%v", a.GetId()),
					a.GetName(),
					status,
					strings.Join(a.GetApprovers(), " "),
				})
			}

			printer.Table(os.Stdout, report)
			return nil
		},
	}

	return cmd
}

func cancelAppealCommand(c *app.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel an appeal",
		Example: heredoc.Doc(`
		$ guardian appeal cancel <appeal-id>
	`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			id := args[0]
			_, err = strconv.ParseUint(id, 10, 32)
			if err != nil {
				return fmt.Errorf("invalid provider id: %v", err)
			}

			_, err = client.CancelAppeal(ctx, &guardianv1beta1.CancelAppealRequest{
				Id: id,
			})
			if err != nil {
				return err
			}

			fmt.Printf("appeal with id `%v` cancelled successfully", id)

			return nil
		},
	}

	return cmd
}
