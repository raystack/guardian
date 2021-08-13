package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func appealsCommand(c *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "appeals",
		Short: "manage appeals",
	}

	cmd.AddCommand(listAppealsCommand(c))
	cmd.AddCommand(createAppealCommand(c))
	cmd.AddCommand(approveApprovalStepCommand(c))
	cmd.AddCommand(rejectApprovalStepCommand(c))

	return cmd
}

func listAppealsCommand(c *config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list appeals",
		RunE: func(cmd *cobra.Command, args []string) error {
			dialTimeoutCtx, dialCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer dialCancel()
			conn, err := createConnection(dialTimeoutCtx, c.Host)
			if err != nil {
				return err
			}
			defer conn.Close()
			client := pb.NewGuardianServiceClient(conn)

			requestTimeoutCtx, requestTimeoutCtxCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer requestTimeoutCtxCancel()
			res, err := client.ListAppeals(requestTimeoutCtx, &pb.ListAppealsRequest{})
			if err != nil {
				return err
			}

			t := getTablePrinter(os.Stdout, []string{"ID", "USER", "RESOURCE ID", "ROLE", "STATUS"})
			for _, a := range res.GetAppeals() {
				t.Append([]string{
					fmt.Sprintf("%v", a.GetId()),
					a.GetUser(),
					fmt.Sprintf("%v", a.GetResourceId()),
					a.GetRole(),
					a.GetStatus(),
				})
			}
			t.Render()
			return nil
		},
	}
}

func createAppealCommand(c *config) *cobra.Command {
	var user string
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

			dialTimeoutCtx, dialCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer dialCancel()
			conn, err := createConnection(dialTimeoutCtx, c.Host)
			if err != nil {
				return err
			}
			defer conn.Close()
			client := pb.NewGuardianServiceClient(conn)

			requestTimeoutCtx, requestTimeoutCtxCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer requestTimeoutCtxCancel()
			res, err := client.CreateAppeal(requestTimeoutCtx, &pb.CreateAppealRequest{
				User: user,
				Resources: []*pb.CreateAppealRequest_Resource{
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

	cmd.Flags().StringVarP(&user, "user", "u", "", "user email")
	cmd.MarkFlagRequired("user")
	cmd.Flags().UintVar(&resourceID, "resource-id", 0, "resource id")
	cmd.MarkFlagRequired("resource-id")
	cmd.Flags().StringVarP(&role, "role", "r", "", "role")
	cmd.MarkFlagRequired("role")
	cmd.Flags().StringVar(&optionsDuration, "options.duration", "", "access duration")

	return cmd
}

func approveApprovalStepCommand(c *config) *cobra.Command {
	var id uint
	var approvalName string

	cmd := &cobra.Command{
		Use:   "approve",
		Short: "approve an approval step",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dialTimeoutCtx, dialCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer dialCancel()
			conn, err := createConnection(dialTimeoutCtx, c.Host)
			if err != nil {
				return err
			}
			defer conn.Close()
			client := pb.NewGuardianServiceClient(conn)

			requestTimeoutCtx, requestTimeoutCtxCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer requestTimeoutCtxCancel()
			_, err = client.UpdateApproval(requestTimeoutCtx, &pb.UpdateApprovalRequest{
				Id:           uint32(id),
				ApprovalName: approvalName,
				Action: &pb.UpdateApprovalRequest_Action{
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

func rejectApprovalStepCommand(c *config) *cobra.Command {
	var id uint
	var approvalName string

	cmd := &cobra.Command{
		Use:   "reject",
		Short: "reject an approval step",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dialTimeoutCtx, dialCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer dialCancel()
			conn, err := createConnection(dialTimeoutCtx, c.Host)
			if err != nil {
				return err
			}
			defer conn.Close()
			client := pb.NewGuardianServiceClient(conn)

			requestTimeoutCtx, requestTimeoutCtxCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer requestTimeoutCtxCancel()
			_, err = client.UpdateApproval(requestTimeoutCtx, &pb.UpdateApprovalRequest{
				Id:           uint32(id),
				ApprovalName: approvalName,
				Action: &pb.UpdateApprovalRequest_Action{
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
