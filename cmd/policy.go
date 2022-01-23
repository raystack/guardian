package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc"
	handlerv1beta1 "github.com/odpf/guardian/api/handler/v1beta1"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"
)

// PolicyCmd is the root command for the policies subcommand.
func PolicyCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "policy",
		Aliases: []string{"policies"},
		Short:   "Manage policies",
		Long: heredoc.Doc(`
			Work with policies.

			Policies are used to define governance rules of the data access.
		`),
		Example: heredoc.Doc(`
			$ guardian policy create
			$ guardian policy edit
			$ guardian policy list
			$ guardian policy view
			$ guardian policy init	
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
	}

	cmd.AddCommand(listPoliciesCmd(c))
	cmd.AddCommand(getPolicyCmd(c, adapter))
	cmd.AddCommand(createPolicyCmd(c, adapter))
	cmd.AddCommand(updatePolicyCmd(c, adapter))
	cmd.AddCommand(initPolicyCmd(c))

	return cmd
}

func listPoliciesCmd(c *app.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List and filter access policies",
		Long: heredoc.Doc(`
			List and filter all available access policies.
		`),
		Example: heredoc.Doc(`
			$ guardian policy list
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Progress("")
			defer spinner.Stop()

			cs := term.NewColorScheme()

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListPolicies(ctx, &guardianv1beta1.ListPoliciesRequest{})
			if err != nil {
				return err
			}

			report := [][]string{}
			index := 0

			policies := res.GetPolicies()

			spinner.Stop()

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

			fmt.Println("\nFor details on a policy, try: guardian policy view <id@version>")
			return nil
		},
	}
}

func getPolicyCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var format, versionFlag string

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View a policy",
		Long: heredoc.Doc(`
			View a policy.

			Display the ID, name, and other information about a policy.
		`),
		Example: heredoc.Doc(`
			$ guardian policy view <policy-id> --version=<policy-version>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var version uint64
			var id string

			spinner := printer.Progress("")
			defer spinner.Stop()

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			policyId := strings.Split(args[0], "@")
			id = policyId[0]

			version, err = getVersion(versionFlag, policyId)
			if err != nil {
				return err
			}

			res, err := client.GetPolicy(ctx, &guardianv1beta1.GetPolicyRequest{
				Id:      id,
				Version: uint32(version),
			})
			if err != nil {
				return err
			}

			p, err := adapter.FromPolicyProto(res.GetPolicy())
			if err != nil {
				return fmt.Errorf("failed to parse policy: %v", err)
			}

			spinner.Stop()

			if err := printer.Text(p, format); err != nil {
				return fmt.Errorf("failed to format policy: %v", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "output", "o", "yaml", "Print output with the selected format")
	cmd.Flags().StringVarP(&versionFlag, "version", "v", "", "Version of the policy")

	return cmd
}

func createPolicyCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new policy",
		Long: heredoc.Doc(`
			Create a new policy from a file.
		`),
		Example: heredoc.Doc(`
			$ guardian policy create --file=<file-path>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Progress("")
			defer spinner.Stop()

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

			res, err := client.CreatePolicy(ctx, &guardianv1beta1.CreatePolicyRequest{
				Policy: policyProto,
			})
			if err != nil {
				return err
			}

			spinner.Stop()

			fmt.Printf("Policy created with id: %v\n", res.GetPolicy().GetId())

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the policy config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func updatePolicyCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a policy",
		Long: heredoc.Doc(`
			Edit an existing policy with a file.
		`),
		Example: heredoc.Doc(`
			$ guardian policy edit --file=<file-path>
		`),

		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Progress("")
			defer spinner.Stop()

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

			policyID := policyProto.GetId()
			_, err = client.UpdatePolicy(ctx, &guardianv1beta1.UpdatePolicyRequest{
				Id:     policyID,
				Policy: policyProto,
			})
			if err != nil {
				return err
			}

			spinner.Stop()

			fmt.Println("Successfully updated policy")

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the policy config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func initPolicyCmd(c *app.CLIConfig) *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Creates a policy template",
		Long: heredoc.Doc(`
			Create a policy template with a given file name.
		`),
		Example: heredoc.Doc(`
			$ guardian policy init --file=<output-name>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			pwd, _ := os.Getwd()
			bytesRead, err := ioutil.ReadFile(pwd + "/spec/policy.yml")
			if err != nil {
				return err
			}

			//Copy all the contents to the desitination file
			err = ioutil.WriteFile(file, bytesRead, 0777)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "File name for the policy config")
	cmd.MarkFlagRequired("file")

	return cmd
}
func getVersion(versionFlag string, policyId []string) (uint64, error) {
	if versionFlag != "" {
		ver, err := strconv.ParseUint(versionFlag, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid policy version: %v", err)
		}
		return ver, nil
	} else {
		if len(policyId) != 2 {
			return 0, fmt.Errorf("policy version is missing")
		}

		ver, err := strconv.ParseUint(policyId[1], 10, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid policy version: %v", err)
		}
		return ver, nil
	}
}
