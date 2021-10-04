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
	"github.com/spf13/cobra"
)

func providersCommand(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "manage providers",
	}

	cmd.AddCommand(listProvidersCommand(c))
	cmd.AddCommand(getProviderCmd(c, adapter))
	cmd.AddCommand(createProviderCommand(c, adapter))
	cmd.AddCommand(updateProviderCommand(c, adapter))

	return cmd
}

func listProvidersCommand(c *app.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list providers",
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

			t := getTablePrinter(os.Stdout, []string{"ID", "TYPE", "URN"})
			for _, p := range res.GetProviders() {
				t.Append([]string{
					fmt.Sprintf("%v", p.GetId()),
					p.GetType(),
					p.GetUrn(),
				})
			}
			t.Render()
			return nil
		},
	}
}

func getProviderCmd(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a provider details",
		Example: heredoc.Doc(`
			$ guardian provider get 1
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

func createProviderCommand(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "register provider configuration",
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

			fmt.Printf("provider created with id: %v", res.GetProvider().GetId())

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func updateProviderCommand(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	var id uint
	var filePath string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update provider configuration",
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

			fmt.Println("provider updated")

			return nil
		},
	}

	cmd.Flags().UintVar(&id, "id", 0, "provider id")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}
