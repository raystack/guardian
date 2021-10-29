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
	"github.com/odpf/guardian/domain"
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

	cmd.AddCommand(listResourcesCmd(c, adapter))
	cmd.AddCommand(getResourceCmd(c, adapter))
	cmd.AddCommand(metadataCmd(c))
	cmd.PersistentFlags().String("format", "", "Print output with specified format (yaml,json,prettyjson)")

	return cmd
}

func listResourcesCmd(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
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

			format := cmd.Flag("format").Value.String()
			if format != "" {
				var resources []*domain.Resource
				for _, r := range res.GetResources() {
					resources = append(resources, adapter.FromResourceProto(r))
				}
				if err := printer.Text(resources, format); err != nil {
					return fmt.Errorf("failed to parse resources: %v", err)
				}
				return nil
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

	cmd.Flags().StringVar(&providerType, "provider-type", "", "Filter resources by provider type")
	cmd.Flags().StringVar(&providerURN, "provider-urn", "", "Filter resources by provider urn")
	cmd.Flags().StringVar(&resourceType, "type", "", "Filter resources by type")
	cmd.Flags().StringVar(&resourceURN, "urn", "", "Filter resources by urn")
	cmd.Flags().StringVar(&name, "name", "", "Filter resources by name")
	cmd.Flags().StringArrayVar(&detailsPaths, "details-paths", nil, "Object paths to filter resources by details")
	cmd.Flags().StringArrayVar(&detailsValues, "details-values", nil, "Values for --details-paths to filter resources by values")
	cmd.Flags().BoolVar(&isDeleted, "show-deleted", false, "Show deleted resources")

	return cmd
}

func getResourceCmd(c *app.CLIConfig, adapter v1.ProtoAdapter) *cobra.Command {
	return &cobra.Command{
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
			format := cmd.Flag("format").Value.String()
			if err := printer.Text(r, format); err != nil {
				return fmt.Errorf("failed to parse resource: %v", err)
			}
			return nil
		},
	}
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
