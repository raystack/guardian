package cmd

import (
	"context"
	"fmt"
	"os"

	v1 "github.com/odpf/guardian/api/handler/v1"
	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/odpf/guardian/domain"
	"github.com/spf13/cobra"
)

func providersCommand(c *config, adapter v1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "manage providers",
	}

	cmd.AddCommand(listProvidersCommand(c))
	cmd.AddCommand(createProviderCommand(c, adapter))
	cmd.AddCommand(updateProviderCommand(c, adapter))

	return cmd
}

func listProvidersCommand(c *config) *cobra.Command {
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

func createProviderCommand(c *config, adapter v1.ProtoAdapter) *cobra.Command {
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

func updateProviderCommand(c *config, adapter v1.ProtoAdapter) *cobra.Command {
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
