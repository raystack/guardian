package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	v1 "github.com/odpf/guardian/api/handler/v1"
	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"
)

//  PolicyCmd is the root command for the policies subcommand.
func PolicyCmd(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage policies",
		Annotations: map[string]string{
			"group:core": "true",
		},
	}

	cmd.AddCommand(listPoliciesCmd(c))
	cmd.AddCommand(createPolicyCmd(c, adapter))
	cmd.AddCommand(updatePolicyCmd(c, adapter))

	return cmd
}

func listPoliciesCmd(c *app.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List and filter access policies.",
		Long: heredoc.Doc(`
			List and filter access policies.

			This command lists and filter all available access policies. 
			Policies are used to define governance rules of the data access.
		`),
		Example: heredoc.Doc(`
			$ guardian policy list
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cs := term.NewColorScheme()

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

			report := [][]string{}
			index := 0

			policies := res.GetPolicies()
			fmt.Printf(" \nShowing %d of %d policies\n \n", len(policies), len(policies))

			report = append(report, []string{"ID", "NAME", "DESCRIPTION", "VERSION", "STEPS"})
			for _, p := range policies {
				report = append(report, []string{
					cs.Greenf("%02d", index),
					fmt.Sprintf("%v", p.GetId()),
					p.GetDescription(),
					fmt.Sprintf("%v", p.GetVersion()),
					fmt.Sprintf("%v", len(p.GetSteps())),
				})
				index++
			}
			printer.Table(os.Stdout, report)

			// fmt.Println("\nFor details on a policy, try: guardian policy view <policy-id>")
			return nil
		},
	}
}

func createPolicyCmd(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new policy",
		Example: heredoc.Doc(`
			$ guardian policy create -f policy.yaml
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
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

			fmt.Printf("Policy created with id: %v\n", res.GetPolicy().GetId())

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "path to the policy config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func updatePolicyCmd(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	var id string
	var filePath string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing policy",
		Example: heredoc.Doc(`
			$ guardian policy update -f policy.yaml
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
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

			fmt.Println("Successfully updated policy")

			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "policy id")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "path to the policy config")
	cmd.MarkFlagRequired("file")

	return cmd
}
