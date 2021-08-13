package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func resourcesCommand(c *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "manage resources",
	}

	cmd.AddCommand(listResourcesCommand(c))
	cmd.AddCommand(metadataCommand(c))

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

func metadataCommand(c *config) *cobra.Command {
	var id uint
	var values []string

	cmd := &cobra.Command{
		Use:   "metadata",
		Short: "manage resource's metadata",
	}

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "store new metadata",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			metadata := map[string]interface{}{}
			for _, a := range values {
				items := strings.Split(a, "=")
				key := items[0]
				value := items[1]

				metadata[key] = value
			}
			metadataProto, err := structpb.NewStruct(metadata)
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

			// TODO: get one resource

			requestTimeoutCtx, requestTimeoutCtxCancel := context.WithTimeout(context.Background(), time.Second*2)
			defer requestTimeoutCtxCancel()
			_, err = client.UpdateResource(requestTimeoutCtx, &pb.UpdateResourceRequest{
				Id: uint32(id),
				Resource: &pb.Resource{
					Details: metadataProto,
				},
			})
			if err != nil {
				return err
			}

			fmt.Println("metadata updated")

			return nil
		},
	}

	setCmd.Flags().StringArrayVar(&values, "values", []string{}, "list of key-value pair. Example: key=value foo=bar")

	cmd.AddCommand(setCmd)
	cmd.PersistentFlags().UintVar(&id, "id", 0, "resource id")
	cmd.MarkPersistentFlagRequired("id")

	return cmd
}
