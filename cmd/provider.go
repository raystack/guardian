package cmd

import (
	"context"
	"fmt"
	"os"

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
			
			Providers are the system for which we intend to manage access.
		`),
		Example: heredoc.Doc(`
			$ guardian provider create -f file.yaml
			$ guardian provider list
			$ guardian provider view 1
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
	}

	cmd.AddCommand(listProvidersCmd(c))
	cmd.AddCommand(getProviderCmd(c, adapter))
	cmd.AddCommand(createProviderCmd(c, adapter))
	cmd.AddCommand(updateProviderCmd(c, adapter))

	return cmd
}

func listProvidersCmd(c *app.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List and filter providers",
		Long: heredoc.Doc(`
			List and filter all registered providers.
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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

func getProviderCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View a provider details",
		Long: heredoc.Doc(`
			View a provider.

			Display the ID, name, and other information about a provider.
		`),
		Example: heredoc.Doc(`
			$ guardian provider view 1
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
		Long: heredoc.Doc(`
			Register a new provider.
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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

			fmt.Printf("Provider created with id: %v", res.GetProvider().GetId())

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func updateProviderCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var id, filePath string
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a provider",
		Long: heredoc.Doc(`
			Edit an existing provider.
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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

			_, err = client.UpdateProvider(ctx, &guardianv1beta1.UpdateProviderRequest{
				Id:     id,
				Config: configProto,
			})
			if err != nil {
				return err
			}

			fmt.Println("Successfully updated provider")

			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "provider id")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}
