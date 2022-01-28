package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	handlerv1beta1 "github.com/odpf/guardian/api/handler/v1beta1"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
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
			_, err = strconv.ParseUint(id, 10, 32)
			if err != nil {
				return fmt.Errorf("invalid provider id: %v", err)
			}

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
