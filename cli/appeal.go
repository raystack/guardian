package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func appealsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "appeal",
		Aliases: []string{"appeals"},
		Short:   "Manage appeals",
		Annotations: map[string]string{
			"group": "core",
		},
		Example: heredoc.Doc(`
			$ guardian appeal create
			$ guardian appeal approve
			$ guardian appeal reject
			$ guardian appeal list
			$ guardian appeal status
			$ guardian appeal revoke
			$ guardian appeal bulk-revoke
			$ guardian appeal cancel
		`),
	}

	cmd.AddCommand(listAppealsCommand())
	cmd.AddCommand(createAppealCommand())
	cmd.AddCommand(approveApprovalStepCommand())
	cmd.AddCommand(rejectApprovalStepCommand())
	cmd.AddCommand(statusAppealCommand())
	cmd.AddCommand(cancelAppealCommand())
	cmd.AddCommand(addApproverCommand())
	cmd.AddCommand(deleteApproverCommand())
	bindFlagsFromConfig(cmd)

	return cmd
}

func listAppealsCommand() *cobra.Command {
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
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListAppeals(cmd.Context(), &guardianv1beta1.ListAppealsRequest{
				Statuses:  statuses,
				Role:      role,
				AccountId: accountID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}

			appeals := res.GetAppeals()
			spinner.Stop()

			fmt.Printf(" \nShowing %d of %d policies\n \n", len(appeals), len(appeals))

			report = append(report, []string{"ID", "USER", "RESOURCE ID", "ROLE", "STATUS"})
			for _, a := range appeals {
				report = append(report, []string{
					term.Greenf("%v", a.GetId()),
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

func createAppealCommand() *cobra.Command {
	var accountID, accountType, resourceID, role, optionsDuration string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new appeal",
		Example: heredoc.Doc(`
			$ guardian appeal create
			$ guardian appeal create --account=<account-id> --type=<account-type> --resource=<resource-id> --role=<role>
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			options := map[string]interface{}{}
			if optionsDuration != "" {
				options["duration"] = optionsDuration
			}
			optionsProto, err := structpb.NewStruct(options)
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.CreateAppeal(cmd.Context(), &guardianv1beta1.CreateAppealRequest{
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

			spinner.Stop()

			appealID := res.GetAppeals()[0].GetId()
			printer.Successf("appeal created with id: %v", appealID)
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

func approveApprovalStepCommand() *cobra.Command {
	var approvalName string

	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve an approval step",
		Example: heredoc.Doc(`
		$ guardian appeal approve <appeal-id> --step=<step-name>
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
			_, err = client.UpdateApproval(cmd.Context(), &guardianv1beta1.UpdateApprovalRequest{
				Id:           id,
				ApprovalName: approvalName,
				Action: &guardianv1beta1.UpdateApprovalRequest_Action{
					Action: "approve",
				},
			})
			if err != nil {
				return err
			}

			spinner.Stop()

			fmt.Printf("appeal with id %v and approval name %v approved successfully", id, approvalName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&approvalName, "step", "s", "", "Name of approval step")
	cmd.MarkFlagRequired("approval-name")

	return cmd
}

func rejectApprovalStepCommand() *cobra.Command {
	var approvalName string

	cmd := &cobra.Command{
		Use:   "reject",
		Short: "Reject an approval step",
		Example: heredoc.Doc(`
		$ guardian appeal reject <appeal-id> --step=<step-name>
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
			_, err = client.UpdateApproval(cmd.Context(), &guardianv1beta1.UpdateApprovalRequest{
				Id:           id,
				ApprovalName: approvalName,
				Action: &guardianv1beta1.UpdateApprovalRequest_Action{
					Action: "reject",
				},
			})
			if err != nil {
				return err
			}

			spinner.Stop()

			fmt.Printf("appeal with id %v and approval name %v rejected successfully", id, approvalName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&approvalName, "step", "s", "", "Name of approval step")
	cmd.MarkFlagRequired("approval-name")

	return cmd
}

func statusAppealCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Approval status of an appeal",
		Example: heredoc.Doc(`
			$ guardian appeal status <appeal-id>
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
			res, err := client.GetAppeal(cmd.Context(), &guardianv1beta1.GetAppealRequest{
				Id: id,
			})
			if err != nil {
				return err
			}

			spinner.Stop()

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

func cancelAppealCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel an appeal",
		Example: heredoc.Doc(`
		$ guardian appeal cancel <appeal-id>
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
			_, err = client.CancelAppeal(cmd.Context(), &guardianv1beta1.CancelAppealRequest{
				Id: id,
			})
			if err != nil {
				return err
			}

			spinner.Stop()

			fmt.Printf("appeal with id `%v` cancelled successfully", id)

			return nil
		},
	}

	return cmd
}

func addApproverCommand() *cobra.Command {
	var appealID, approvalID, email string

	cmd := &cobra.Command{
		Use:   "add-approver",
		Short: "Add a new approver to an approval step",
		Example: heredoc.Doc(`
			$ guardian appeal add-approver --appeal-id=<appeal-id> --approval-id=<approval-id> --email=<new-approver-email>
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			if _, err := client.AddApprover(cmd.Context(), &guardianv1beta1.AddApproverRequest{
				AppealId:   appealID,
				ApprovalId: approvalID,
				Email:      email,
			}); err != nil {
				return err
			}
			spinner.Stop()

			fmt.Printf("%q added to the approval\n", email)
			return nil
		},
	}

	cmd.Flags().StringVar(&appealID, "appeal-id", "", "Appeal ID")
	cmd.Flags().StringVar(&approvalID, "approval-id", "", "Approval ID or approval name")
	cmd.Flags().StringVar(&email, "email", "", "New approver email")
	cmd.MarkFlagRequired("appeal-id")
	cmd.MarkFlagRequired("approval-id")
	cmd.MarkFlagRequired("email")

	return cmd
}

func deleteApproverCommand() *cobra.Command {
	var appealID, approvalID, email string

	cmd := &cobra.Command{
		Use:   "delete-approver",
		Short: "Remove an existing approver from an approval step",
		Example: heredoc.Doc(`
			$ guardian appeal delete-approver --appeal-id=<appeal-id> --approval-id=<approval-id> --email=<new-approver-email>
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			if _, err := client.DeleteApprover(cmd.Context(), &guardianv1beta1.DeleteApproverRequest{
				AppealId:   appealID,
				ApprovalId: approvalID,
				Email:      email,
			}); err != nil {
				return err
			}
			spinner.Stop()

			fmt.Printf("%q removed from the approval\n", email)
			return nil
		},
	}

	cmd.Flags().StringVar(&appealID, "appeal-id", "", "Appeal ID")
	cmd.Flags().StringVar(&approvalID, "approval-id", "", "Approval ID or approval name")
	cmd.Flags().StringVar(&email, "email", "", "New approver email")
	cmd.MarkFlagRequired("appeal-id")
	cmd.MarkFlagRequired("approval-id")
	cmd.MarkFlagRequired("email")

	return cmd
}
