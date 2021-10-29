package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc"
	v1 "github.com/odpf/guardian/api/handler/v1"
	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/odpf/guardian/app"
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func ResourceCmd(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource",
		Aliases: []string{"resources"},
		Short:   "Manage resources",
		Example: heredoc.Doc(`
			$ guardian resource list
			$ guardian resource metadata set --id=1 key=value
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
	}

	cmd.AddCommand(listResourcesCmd(c))
	cmd.AddCommand(getResourceCmd(c, adapter))
	cmd.AddCommand(metadataCmd(c))

	return cmd
}

func listResourcesCmd(c *app.CLIConfig) *cobra.Command {
	var providerType, providerURN, resourceType, resourceURN, name string
	var isDeleted bool
	var detailsPaths, detailsValues []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resources",
		Example: heredoc.Doc(`
			$ guardian resource list
			$ guardian resource list --provider-type=bigquery --type=dataset
			$ guardian resource list --details-paths=foo.bar --details-values=123 
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

			req := &pb.ListResourcesRequest{
				ProviderType:  providerType,
				ProviderUrn:   providerURN,
				Type:          resourceType,
				Urn:           resourceURN,
				Name:          name,
				IsDeleted:     isDeleted,
				DetailsPaths:  detailsPaths,
				DetailsValues: detailsValues,
			}
			res, err := client.ListResources(ctx, req)
			if err != nil {
				return err
			}

			report := [][]string{}

			resources := res.GetResources()
			fmt.Printf(" \nShowing %d of %d resources\n \n", len(resources), len(resources))

			report = append(report, []string{"ID", "PROVIDER", "TYPE", "URN", "NAME"})
			for _, r := range resources {
				report = append(report, []string{
					fmt.Sprintf("%v", r.GetId()),
					fmt.Sprintf("%s\n%s", r.GetProviderType(), r.GetProviderUrn()),
					r.GetType(),
					r.GetUrn(),
					r.GetName(),
				})
			}
			printer.Table(os.Stdout, report)
			return nil
		},
	}

	cmd.Flags().StringVar(&providerType, "provider-type", "", "filter resources by provider type")
	cmd.Flags().StringVar(&providerURN, "provider-urn", "", "filter resources by provider urn")
	cmd.Flags().StringVar(&resourceType, "type", "", "filter resources by type")
	cmd.Flags().StringVar(&resourceURN, "urn", "", "filter resources by urn")
	cmd.Flags().StringVar(&name, "name", "", "filter resources by name")
	cmd.Flags().BoolVar(&isDeleted, "show-deleted", false, "show deleted resources")
	cmd.Flags().StringArrayVar(&detailsPaths, "details-paths", nil, "")
	cmd.Flags().StringArrayVar(&detailsValues, "details-values", nil, "")

	return cmd
}

func getResourceCmd(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a resource details",
		Example: heredoc.Doc(`
			$ guardian resource get 1
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
				return fmt.Errorf("invalid resource id: %v", err)
			}

			res, err := client.GetResource(ctx, &pb.GetResourceRequest{
				Id: uint32(id),
			})
			if err != nil {
				return err
			}

			r := adapter.FromResourceProto(res)
			if err := printer.Text(r, format); err != nil {
				return fmt.Errorf("failed to parse resource: %v", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "yaml", "Print output with the selected format")

	return cmd
}

func metadataCmd(c *app.CLIConfig) *cobra.Command {
	var id uint
	var values []string

	cmd := &cobra.Command{
		Use:   "metadata",
		Short: "Manage resource's metadata",
	}

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Store new metadata",
		Example: heredoc.Doc(`
			$ guardian resource metadata set values foo=bar
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
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

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			// TODO: get one resource

			_, err = client.UpdateResource(ctx, &pb.UpdateResourceRequest{
				Id: uint32(id),
				Resource: &pb.Resource{
					Details: metadataProto,
				},
			})
			if err != nil {
				return err
			}

			fmt.Println("Successfully updated metadata")

			return nil
		},
	}

	setCmd.Flags().StringArrayVar(&values, "values", []string{}, "list of key-value pair. Example: key=value foo=bar")

	cmd.AddCommand(setCmd)
	cmd.PersistentFlags().UintVar(&id, "id", 0, "resource id")
	cmd.MarkPersistentFlagRequired("id")

	return cmd
}
