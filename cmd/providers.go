package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	v1 "github.com/odpf/guardian/api/handler/v1"
	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/odpf/guardian/domain"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
			dialTimeoutCtx, dialCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer dialCancel()
			conn, err := createConnection(dialTimeoutCtx, c.Host)
			if err != nil {
				return err
			}
			defer conn.Close()
			client := pb.NewGuardianServiceClient(conn)

			requestTimeoutCtx, requestTimeoutCtxCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer requestTimeoutCtxCancel()
			res, err := client.ListProviders(requestTimeoutCtx, &pb.ListProvidersRequest{})
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
			b, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			var providerConfig domain.ProviderConfig
			switch filepath.Ext(filePath) {
			case ".json":
				if err := json.Unmarshal(b, &providerConfig); err != nil {
					return err
				}
			case ".yaml", ".yml":
				if err := yaml.Unmarshal(b, &providerConfig); err != nil {
					return err
				}
			default:
				return errors.New("unsupported file type")
			}
			configProto, err := adapter.ToProviderConfigProto(&providerConfig)
			if err != nil {
				return err
			}

			dialTimeoutCtx, dialCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer dialCancel()
			conn, err := createConnection(dialTimeoutCtx, c.Host)
			if err != nil {
				return err
			}
			defer conn.Close()
			client := pb.NewGuardianServiceClient(conn)

			requestTimeoutCtx, requestTimeoutCtxCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer requestTimeoutCtxCancel()
			res, err := client.CreateProvider(requestTimeoutCtx, &pb.CreateProviderRequest{
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
	var filePath string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update provider configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			var providerConfig domain.ProviderConfig
			switch filepath.Ext(filePath) {
			case ".json":
				if err := json.Unmarshal(b, &providerConfig); err != nil {
					return err
				}
			case ".yaml", ".yml":
				if err := yaml.Unmarshal(b, &providerConfig); err != nil {
					return err
				}
			default:
				return errors.New("unsupported file type")
			}
			configProto, err := adapter.ToProviderConfigProto(&providerConfig)
			if err != nil {
				return err
			}

			dialTimeoutCtx, dialCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer dialCancel()
			conn, err := createConnection(dialTimeoutCtx, c.Host)
			if err != nil {
				return err
			}
			defer conn.Close()
			client := pb.NewGuardianServiceClient(conn)

			requestTimeoutCtx, requestTimeoutCtxCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer requestTimeoutCtxCancel()
			_, err = client.UpdateProvider(requestTimeoutCtx, &pb.UpdateProviderRequest{
				Config: configProto,
			})
			if err != nil {
				return err
			}

			fmt.Println("provider updated")

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}
