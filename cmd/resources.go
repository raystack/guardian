package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/spf13/cobra"
)

func resourcesCommand(c *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "manage resources",
	}

	cmd.AddCommand(listResourcesCommand(c))

	return cmd
}

func listResourcesCommand(c *config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list resources",
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
			res, err := client.ListResources(requestTimeoutCtx, &pb.ListResourcesRequest{})
			if err != nil {
				return err
			}

			t := getTablePrinter(os.Stdout, []string{"ID", "PROVIDER", "TYPE", "URN", "NAME"})
			for _, r := range res.GetResources() {
				t.Append([]string{
					fmt.Sprintf("%v", r.GetId()),
					fmt.Sprintf("%s\n%s", r.GetProviderType(), r.GetProviderUrn()),
					r.GetType(),
					r.GetUrn(),
					r.GetName(),
				})
			}
			t.Render()
			return nil
		},
	}
}
