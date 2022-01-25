package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/MakeNowJust/heredoc"
	handlerv1beta1 "github.com/odpf/guardian/api/handler/v1beta1"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/printer"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func ProviderCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
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
			"group:core": "true",
		},
	}

	cmd.AddCommand(listProvidersCmd(c))
	cmd.AddCommand(viewProviderCmd(c, adapter))
	cmd.AddCommand(createProviderCmd(c, adapter))
	cmd.AddCommand(editProviderCmd(c, adapter))
	cmd.AddCommand(planProviderCmd(c, adapter))
	cmd.AddCommand(applyProviderCmd(c, adapter))
	cmd.AddCommand(initProviderCmd(c))

	return cmd
}

func listProvidersCmd(c *app.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List and filter providers",
		Long:  "List and filter all registered providers.",
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListProviders(ctx, &guardianv1beta1.ListProvidersRequest{})
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

func viewProviderCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
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
			"group:core": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			id := args[0]
			res, err := client.GetProvider(ctx, &guardianv1beta1.GetProviderRequest{
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

			if err := printer.Text(p, format); err != nil {
				return fmt.Errorf("failed to format provider: %v", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "output", "o", "yaml", "Print output with the selected format")

	return cmd
}

func createProviderCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register a new provider",
		Long:  "Register a new provider.",
		Example: heredoc.Doc(`
			$ guardian provider create --file <file-path>
		`),
		Annotations: map[string]string{
			"group:core": "true",
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

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.CreateProvider(ctx, &guardianv1beta1.CreateProviderRequest{
				Config: configProto,
			})
			if err != nil {
				return err
			}

			spinner.Stop()

			fmt.Printf("Provider created with id: %v", res.GetProvider().GetId())

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func editProviderCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a provider",
		Long:  "Edit an existing provider.",
		Example: heredoc.Doc(`
			$ guardian provider edit <provider-id> --file <file-path>
		`),
		Annotations: map[string]string{
			"group:core": "true",
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

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			id := args[0]
			_, err = client.UpdateProvider(ctx, &guardianv1beta1.UpdateProviderRequest{
				Id:     id,
				Config: configProto,
			})
			if err != nil {
				return err
			}

			spinner.Stop()

			fmt.Println("Successfully updated provider")

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func initProviderCmd(c *app.CLIConfig) *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Creates a provider template",
		Long:  "Create a provider template with a given file name.",
		Example: heredoc.Doc(`
			$ guardian provider init --file=<output-name>
		`),
		Annotations: map[string]string{
			"group:core": "true",
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

func applyProviderCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var filePath string
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
			"group:core": "true",
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

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			pType := configProto.GetType()
			pUrn := configProto.GetUrn()

			listRes, err := client.ListProviders(ctx, &guardianv1beta1.ListProvidersRequest{}) // TODO: filter by type & urn
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
				res, err := client.CreateProvider(ctx, &guardianv1beta1.CreateProviderRequest{
					Config: configProto,
				})
				if err != nil {
					return err
				}

				spinner.Stop()
				fmt.Printf("Provider created with id: %v", res.GetProvider().GetId())
			} else {
				_, err = client.UpdateProvider(ctx, &guardianv1beta1.UpdateProviderRequest{
					Id:     providerID,
					Config: configProto,
				})
				if err != nil {
					return err
				}

				spinner.Stop()
				fmt.Println("Successfully updated provider")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func planProviderCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
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
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var newProvider domain.ProviderConfig
			if err := parseFile(filePath, &newProvider); err != nil {
				return err
			}

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			pType := newProvider.Type
			pUrn := newProvider.URN

			listRes, err := client.ListProviders(ctx, &guardianv1beta1.ListProvidersRequest{}) // TODO: filter by type & urn
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
				getRes, err := client.GetProvider(ctx, &guardianv1beta1.GetProviderRequest{
					Id: providerID,
				})
				if err != nil {
					return err
				}

				pc, err := adapter.FromProviderConfigProto(getRes.GetProvider().GetConfig())
				if err != nil {
					return fmt.Errorf("unable to parse existing provider: %w", err)
				}
				existingProvider = pc
			}

			existingProviderYaml, err := yaml.Marshal(existingProvider)
			if err != nil {
				return fmt.Errorf("failed to marshal existing provider: %w", err)
			}
			newProviderYaml, err := yaml.Marshal(newProvider)
			if err != nil {
				return fmt.Errorf("failed to marshal new provider: %w", err)
			}

			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(string(existingProviderYaml), string(newProviderYaml), false)

			spinner.Stop()
			fmt.Println(dmp.DiffPrettyText(diffs))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}
