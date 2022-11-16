package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/MakeNowJust/heredoc"
	handlerv1beta1 "github.com/odpf/guardian/api/handler/v1beta1"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func ProviderCmd(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "provider",
		Aliases: []string{"providers"},
		Short:   "Manage providers",
		Long: heredoc.Doc(`
			Work with providers.
			
			Providers are the system for which we intend to manage access.`),
		Example: heredoc.Doc(`
			$ guardian provider create
			$ guardian provider list
			$ guardian provider view
			$ guardian provider edit
			$ guardian provider init
		`),
		Annotations: map[string]string{
			"group": "core",
		},
	}

	cmd.AddCommand(listProvidersCmd())
	cmd.AddCommand(viewProviderCmd(adapter))
	cmd.AddCommand(createProviderCmd(adapter))
	cmd.AddCommand(editProviderCmd(adapter))
	cmd.AddCommand(planProviderCmd(adapter))
	cmd.AddCommand(applyProviderCmd(adapter))
	cmd.AddCommand(initProviderCmd())
	bindFlagsFromConfig(cmd)

	return cmd
}

func listProvidersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List and filter providers",
		Long:  "List and filter all registered providers.",
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListProviders(cmd.Context(), &guardianv1beta1.ListProvidersRequest{})
			if err != nil {
				return err
			}

			report := [][]string{}

			providers := res.GetProviders()

			spinner.Stop()

			fmt.Printf(" \nShowing %d of %d providers\n \n", len(providers), len(providers))

			report = append(report, []string{"ID", "TYPE", "URN"})

			for _, p := range providers {
				report = append(report, []string{
					fmt.Sprintf("%v", p.GetId()),
					p.GetType(),
					p.GetUrn(),
				})
			}
			printer.Table(os.Stdout, report)

			fmt.Println("\nFor details on a provider, try: guardian provider view <id>")
			return nil
		},
	}
}

func viewProviderCmd(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View a provider details",
		Long: heredoc.Doc(`
			View a provider.

			Display the ID, name, and other information about a provider.`),
		Example: heredoc.Doc(`
			$ guardian provider view <provider-id>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
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
			res, err := client.GetProvider(cmd.Context(), &guardianv1beta1.GetProviderRequest{
				Id: id,
			})
			if err != nil {
				return err
			}

			p, err := adapter.FromProviderProto(res.GetProvider())
			if err != nil {
				return fmt.Errorf("failed to parse provider: %v", err)
			}

			spinner.Stop()

			if err := printer.File(p, format); err != nil {
				return fmt.Errorf("failed to format provider: %v", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "output", "o", "yaml", "Print output with the selected format")

	return cmd
}

func createProviderCmd(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var filePath string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register a new provider",
		Long:  "Register a new provider.",
		Example: heredoc.Doc(`
			$ guardian provider create --file <file-path>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var providerConfig domain.ProviderConfig
			if err := parseFile(filePath, &providerConfig); err != nil {
				return err
			}

			configProto, err := adapter.ToProviderConfigProto(&providerConfig)
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.CreateProvider(cmd.Context(), &guardianv1beta1.CreateProviderRequest{
				Config: configProto,
				DryRun: dryRun,
			})
			if err != nil {
				return err
			}

			spinner.Stop()

			msg := "Provider created with id: %v"
			if dryRun {
				msg += " (dry run)"
			}

			fmt.Printf(msg+"\n", res.GetProvider().GetId())

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "if true the provider will not be created")
	cmd.MarkFlagRequired("file")

	return cmd
}

func editProviderCmd(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var filePath string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a provider",
		Long:  "Edit an existing provider.",
		Example: heredoc.Doc(`
			$ guardian provider edit <provider-id> --file <file-path>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var providerConfig domain.ProviderConfig
			if err := parseFile(filePath, &providerConfig); err != nil {
				return err
			}

			configProto, err := adapter.ToProviderConfigProto(&providerConfig)
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			id := args[0]
			_, err = client.UpdateProvider(cmd.Context(), &guardianv1beta1.UpdateProviderRequest{
				Id:     id,
				Config: configProto,
				DryRun: dryRun,
			})
			if err != nil {
				return err
			}

			spinner.Stop()

			msg := "Successfully updated provider"
			if dryRun {
				msg += " (dry run)"
			}

			fmt.Println(msg)

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "if true the provider will not be updated")
	cmd.MarkFlagRequired("file")

	return cmd
}

func initProviderCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Creates a provider template",
		Long:  "Create a provider template with a given file name.",
		Example: heredoc.Doc(`
			$ guardian provider init --file=<output-name>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			pwd, _ := os.Getwd()
			bytesRead, err := ioutil.ReadFile(pwd + "/spec/provider.yml")
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

func applyProviderCmd(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var filePath string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a provider",
		Long: heredoc.Doc(`
			Create or edit a provider from a file.
		`),
		Example: heredoc.Doc(`
			$ guardian provider apply --file <file-path>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var providerConfig domain.ProviderConfig
			if err := parseFile(filePath, &providerConfig); err != nil {
				return err
			}

			configProto, err := adapter.ToProviderConfigProto(&providerConfig)
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			pType := configProto.GetType()
			pUrn := configProto.GetUrn()

			listRes, err := client.ListProviders(cmd.Context(), &guardianv1beta1.ListProvidersRequest{}) // TODO: filter by type & urn
			if err != nil {
				return err
			}
			providerID := ""
			for _, p := range listRes.GetProviders() {
				if p.GetType() == pType && p.GetUrn() == pUrn {
					providerID = p.GetId()
				}
			}

			if providerID == "" {
				res, err := client.CreateProvider(cmd.Context(), &guardianv1beta1.CreateProviderRequest{
					Config: configProto,
					DryRun: dryRun,
				})
				if err != nil {
					return err
				}

				spinner.Stop()

				msg := "Provider created with id: %v"
				if dryRun {
					msg += " (dry run)"
				}

				fmt.Printf(msg+"\n", res.GetProvider().GetId())
			} else {
				_, err = client.UpdateProvider(cmd.Context(), &guardianv1beta1.UpdateProviderRequest{
					Id:     providerID,
					Config: configProto,
					DryRun: dryRun,
				})
				if err != nil {
					return err
				}

				spinner.Stop()

				msg := "Successfully updated provider"
				if dryRun {
					msg += " (dry run)"
				}
				fmt.Println(msg)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "if true the provider will not be created or updated")
	cmd.MarkFlagRequired("file")

	return cmd
}

func planProviderCmd(adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Show changes from the new provider",
		Long: heredoc.Doc(`
			Show changes from the new provider. This will not actually apply the provider config.
		`),
		Example: heredoc.Doc(`
			$ guardian provider plan --file=<file-path>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var newProvider domain.ProviderConfig
			if err := parseFile(filePath, &newProvider); err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			pType := newProvider.Type
			pUrn := newProvider.URN

			listRes, err := client.ListProviders(cmd.Context(), &guardianv1beta1.ListProvidersRequest{}) // TODO: filter by type & urn
			if err != nil {
				return err
			}
			providerID := ""
			for _, p := range listRes.GetProviders() {
				if p.GetType() == pType && p.GetUrn() == pUrn {
					providerID = p.GetId()
				}
			}

			var existingProvider *domain.ProviderConfig
			if providerID != "" {
				getRes, err := client.GetProvider(cmd.Context(), &guardianv1beta1.GetProviderRequest{
					Id: providerID,
				})
				if err != nil {
					return err
				}

				existingProvider = adapter.FromProviderConfigProto(getRes.GetProvider().GetConfig())
			}

			existingProvider.Credentials = nil
			newProvider.Credentials = nil
			// TODO: show decrypted credentials value instead of omitting them

			existingProviderYaml, err := yaml.Marshal(existingProvider)
			if err != nil {
				return fmt.Errorf("failed to marshal existing provider: %w", err)
			}
			newProviderYaml, err := yaml.Marshal(newProvider)
			if err != nil {
				return fmt.Errorf("failed to marshal new provider: %w", err)
			}

			diffs := diff(string(existingProviderYaml), string(newProviderYaml))

			spinner.Stop()
			fmt.Println(diffs)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}
