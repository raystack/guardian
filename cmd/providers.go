package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	v1 "github.com/odpf/guardian/api/handler/v1"
	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
)

func ProviderCmd(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "provider",
		Aliases: []string{"providers"},
		Short:   "Manage providers",
		Long: heredoc.Doc(`
			Work with providers.
			
			Providers are the system for which we intend to mange access.
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
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListProviders(ctx, &pb.ListProvidersRequest{})
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

func getProviderCmd(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
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

			id, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid provider id: %v", err)
			}

			res, err := client.GetProvider(ctx, &pb.GetProviderRequest{
				Id: uint32(id),
			})
			if err != nil {
				return err
			}

			p, err := adapter.FromProviderProto(res)
			if err != nil {
				return fmt.Errorf("failed to parse provider: %v", err)
			}

			formattedResult, err := outputFormat(p, format)
			if err != nil {
				return fmt.Errorf("failed to format provider: %v", err)
			}

			fmt.Println(formattedResult)
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "yaml", "Print output with the selected format")

	return cmd
}

func createProviderCmd(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register a new provider",
		Long: heredoc.Doc(`
			Register a new provider.
		`),
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

			res, err := client.CreateProvider(ctx, &pb.CreateProviderRequest{
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

func updateProviderCmd(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	var id uint
	var filePath string
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a provider",
		Long: heredoc.Doc(`
			Edit an existing provider.
		`),
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

			_, err = client.UpdateProvider(ctx, &pb.UpdateProviderRequest{
				Id:     uint32(id),
				Config: configProto,
			})
			if err != nil {
				return err
			}

			fmt.Println("Successfully updated provider")

			return nil
		},
	}

	cmd.Flags().UintVar(&id, "id", 0, "provider id")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}
