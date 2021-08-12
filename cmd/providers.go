package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	v1 "github.com/odpf/guardian/api/handler/v1"
	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/spf13/cobra"
)

func providersCommand(c *config, adapter v1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "manage providers",
	}

	cmd.AddCommand(listProvidersCommand(c))

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
