package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/spf13/cobra"
)

func appealsCommand(c *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "appeals",
		Short: "manage appeals",
	}

	cmd.AddCommand(listAppealsCommand(c))

	return cmd
}

func listAppealsCommand(c *config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list appeals",
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
			res, err := client.ListAppeals(requestTimeoutCtx, &pb.ListAppealsRequest{})
			if err != nil {
				return err
			}

			t := getTablePrinter(os.Stdout, []string{"ID", "USER", "RESOURCE ID", "ROLE", "STATUS"})
			for _, a := range res.GetAppeals() {
				t.Append([]string{
					fmt.Sprintf("%v", a.GetId()),
					a.GetUser(),
					fmt.Sprintf("%v", a.GetResourceId()),
					a.GetRole(),
					a.GetStatus(),
				})
			}
			t.Render()
			return nil
		},
	}
}
