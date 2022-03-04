package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	handlerv1beta1 "github.com/odpf/guardian/api/handler/v1beta1"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

// PolicyCmd is the root command for the policies subcommand.
func PolicyCmd(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "policy",
		Aliases: []string{"policies"},
		Short:   "Manage policies",
		Long: heredoc.Doc(`
			Work with policies.

			Policies are used to define governance rules of the data access.`),
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

	cmd.AddCommand(listPoliciesCmd())
	cmd.AddCommand(getPolicyCmd(adapter))
	cmd.AddCommand(createPolicyCmd(adapter))
	cmd.AddCommand(updatePolicyCmd(adapter))
	cmd.AddCommand(planPolicyCmd(adapter))
	cmd.AddCommand(applyPolicyCmd(adapter))
	cmd.AddCommand(initPolicyCmd())

	return cmd
}

func listPoliciesCmd() *cobra.Command {
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
			spinner := printer.Spin("")
			defer spinner.Stop()

			cs := term.NewColorScheme()

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListPolicies(cmd.Context(), &guardianv1beta1.ListPoliciesRequest{})
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

func getPolicyCmd(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
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

			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd)
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

			res, err := client.GetPolicy(cmd.Context(), &guardianv1beta1.GetPolicyRequest{
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

func createPolicyCmd(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
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
			spinner := printer.Spin("")
			defer spinner.Stop()

			var policy domain.Policy
			if err := parseFile(filePath, &policy); err != nil {
				return err
			}

			policyProto, err := adapter.ToPolicyProto(&policy)
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.CreatePolicy(cmd.Context(), &guardianv1beta1.CreatePolicyRequest{
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

func updatePolicyCmd(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
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
			spinner := printer.Spin("")
			defer spinner.Stop()

			var policy domain.Policy
			if err := parseFile(filePath, &policy); err != nil {
				return err
			}

			policyProto, err := adapter.ToPolicyProto(&policy)
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			policyID := policyProto.GetId()
			_, err = client.UpdatePolicy(cmd.Context(), &guardianv1beta1.UpdatePolicyRequest{
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

func initPolicyCmd() *cobra.Command {
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

func applyPolicyCmd(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a policy config",
		Long: heredoc.Doc(`
			Create or edit a policy from a file.
		`),
		Example: heredoc.Doc(`
			$ guardian policy apply --file=<file-path>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var policy domain.Policy
			if err := parseFile(filePath, &policy); err != nil {
				return err
			}

			policyProto, err := adapter.ToPolicyProto(&policy)
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			policyID := policyProto.GetId()
			_, err = client.GetPolicy(cmd.Context(), &guardianv1beta1.GetPolicyRequest{
				Id: policyID,
			})
			policyExists := true
			if err != nil {
				if e, ok := status.FromError(err); ok && e.Code() == codes.NotFound {
					policyExists = false
				} else {
					return err
				}
			}

			if policyExists {
				res, err := client.UpdatePolicy(cmd.Context(), &guardianv1beta1.UpdatePolicyRequest{
					Id:     policyID,
					Policy: policyProto,
				})
				if err != nil {
					return err
				}
				spinner.Stop()

				fmt.Printf("Policy updated to version: %v\n", res.GetPolicy().GetVersion())
			} else {
				res, err := client.CreatePolicy(cmd.Context(), &guardianv1beta1.CreatePolicyRequest{
					Policy: policyProto,
				})
				if err != nil {
					return err
				}
				spinner.Stop()

				fmt.Printf("Policy created with id: %v\n", res.GetPolicy().GetId())
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the policy config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func planPolicyCmd(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Show changes from the new policy",
		Long: heredoc.Doc(`
			Show changes from the new policy. This will not actually apply the policy config.
		`),
		Example: heredoc.Doc(`
			$ guardian policy plan --file=<file-path>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var newPolicy domain.Policy
			if err := parseFile(filePath, &newPolicy); err != nil {
				return err
			}

			policyProto, err := adapter.ToPolicyProto(&newPolicy)
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			policyID := policyProto.GetId()
			res, err := client.GetPolicy(cmd.Context(), &guardianv1beta1.GetPolicyRequest{
				Id: policyID,
			})
			if err != nil {
				return err
			}

			existingPolicy, err := adapter.FromPolicyProto(res.GetPolicy())
			if err != nil {
				return fmt.Errorf("unable to parse existing policy: %w", err)
			}
			if existingPolicy != nil {
				newPolicy.Version = existingPolicy.Version + 1
				newPolicy.CreatedAt = existingPolicy.CreatedAt
			} else {
				newPolicy.Version = 1
				newPolicy.CreatedAt = time.Now()
			}
			newPolicy.UpdatedAt = time.Now()

			existingPolicyYaml, err := yaml.Marshal(existingPolicy)
			if err != nil {
				return fmt.Errorf("failed to marshal existing policy: %w", err)
			}
			newPolicyYaml, err := yaml.Marshal(newPolicy)
			if err != nil {
				return fmt.Errorf("failed to marshal new policy: %w", err)
			}

			diffs := diff(string(existingPolicyYaml), string(newPolicyYaml))

			spinner.Stop()
			fmt.Println(diffs)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the policy config")
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
