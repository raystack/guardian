package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	v1 "github.com/odpf/guardian/api/handler/v1"
	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/domain"
	"github.com/spf13/cobra"
)

func policiesCommand(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policies",
		Short: "manage policies",
	}

	cmd.AddCommand(listPoliciesCommand(c))
	cmd.AddCommand(createPolicyCommand(c, adapter))
	cmd.AddCommand(updatePolicyCommand(c, adapter))

	return cmd
}

func listPoliciesCommand(c *app.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListPolicies(ctx, &pb.ListPoliciesRequest{})
			if err != nil {
				return err
			}

			t := getTablePrinter(os.Stdout, []string{"ID", "VERSION", "DESCRIPTION", "STEPS"})
			for _, p := range res.GetPolicies() {
				var stepNames []string
				for _, s := range p.GetSteps() {
					stepNames = append(stepNames, s.GetName())
				}
				t.Append([]string{
					fmt.Sprintf("%v", p.GetId()),
					fmt.Sprintf("%v", p.GetVersion()),
					p.GetDescription(),
					strings.Join(stepNames, ","),
				})
			}
			t.Render()
			return nil
		},
	}
}

func createPolicyCommand(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			var policy domain.Policy
			if err := parseFile(filePath, &policy); err != nil {
				return err
			}

			policyProto, err := adapter.ToPolicyProto(&policy)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.CreatePolicy(ctx, &pb.CreatePolicyRequest{
				Policy: policyProto,
			})
			if err != nil {
				return err
			}

			fmt.Printf("policy created with id: %v", res.GetPolicy().GetId())

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "path to the policy config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func updatePolicyCommand(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	var id string
	var filePath string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			var policy domain.Policy
			if err := parseFile(filePath, &policy); err != nil {
				return err
			}

			policyProto, err := adapter.ToPolicyProto(&policy)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			policyID := id
			if policyID == "" {
				policyID = policyProto.GetId()
			}
			_, err = client.UpdatePolicy(ctx, &pb.UpdatePolicyRequest{
				Id:     policyID,
				Policy: policyProto,
			})
			if err != nil {
				return err
			}

			fmt.Println("policy updated")

			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "policy id")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "path to the policy config")
	cmd.MarkFlagRequired("file")

	return cmd
}
