package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc"
	handlerv1beta1 "github.com/odpf/guardian/api/handler/v1beta1"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func ResourceCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource",
		Aliases: []string{"resources"},
		Short:   "Manage resources",
		Example: heredoc.Doc(`
			$ guardian resource list
			$ guardian resource view 1
			$ guardian resource set --id=1 key=value
			$ guardian resource metadata-set
			$ guardian resource metadata-get
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
	}

	cmd.AddCommand(listResourcesCmd(c, adapter))
	cmd.AddCommand(getResourceCmd(c, adapter))
	cmd.AddCommand(metadataSetCmd(c))
	cmd.AddCommand(metadataGetCmd(c))
	cmd.PersistentFlags().StringP("output", "o", "", "Print output with specified format (yaml,json,prettyjson)")

	return cmd
}

func listResourcesCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var providerType, providerURN, resourceType, resourceURN, name, format string
	var isDeleted bool
	var details []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resources",
		Example: heredoc.Doc(`
			$ guardian resource list
			$ guardian resource list --provider-type=bigquery --type=dataset
			$ guardian resource list --details=key1.key2:value --details=key1.key3:value
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

			req := &guardianv1beta1.ListResourcesRequest{
				ProviderType: providerType,
				ProviderUrn:  providerURN,
				Type:         resourceType,
				Urn:          resourceURN,
				Name:         name,
				IsDeleted:    isDeleted,
				Details:      details,
			}
			res, err := client.ListResources(ctx, req)
			if err != nil {
				return err
			}

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

	cmd.Flags().StringVarP(&providerType, "provider-type", "T", "", "Filter by provider type")
	cmd.Flags().StringVarP(&providerURN, "provider-urn", "U", "", "Filter by provider urn")
	cmd.Flags().StringVarP(&resourceType, "type", "t", "", "Filter by type")
	cmd.Flags().StringVarP(&resourceURN, "urn", "u", "", "Filter by urn")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Filter by name")
	cmd.Flags().StringArrayVarP(&details, "details", "d", nil, "Filter by details object values. Example: --details=key1.key2:value")
	cmd.Flags().BoolVarP(&isDeleted, "deleted", "D", false, "Show deleted resources")
	cmd.Flags().StringVarP(&format, "format", "f", "", "Format of output - json yaml prettyjson etc")

	return cmd
}

func getResourceCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View a resource details",
		Example: heredoc.Doc(`
			$ guardian resource view 1
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

			res, err := client.GetResource(ctx, &guardianv1beta1.GetResourceRequest{
				Id: uint32(id),
			})
			if err != nil {
				return err
			}

			if format != "" {
				r := adapter.FromResourceProto(res.GetResource())
				if err := printer.Text(r, format); err != nil {
					return fmt.Errorf("failed to parse resources: %v", err)
				}
				return nil
			}

			report := [][]string{}
			r := res.GetResource()

			report = append(report, []string{"ID", "PROVIDER", "TYPE", "URN", "NAME"})

			report = append(report, []string{
				fmt.Sprintf("%v", r.GetId()),
				fmt.Sprintf("%s\n%s", r.GetProviderType(), r.GetProviderUrn()),
				r.GetType(),
				r.GetUrn(),
				r.GetName(),
			})

			printer.Table(os.Stdout, report)
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "", "Format of output - json yaml prettyjson etc")
	return cmd
}

func metadataSetCmd(c *app.CLIConfig) *cobra.Command {
	var id uint
	var values []string

	cmd := &cobra.Command{
		Use:   "metadata-set",
		Short: "Store new metadata for a resource",
		Example: heredoc.Doc(`
			$ guardian resource metadata-set --id=<resource-id> --values foo=bar
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

			_, err = client.UpdateResource(ctx, &guardianv1beta1.UpdateResourceRequest{
				Id: uint32(id),
				Resource: &guardianv1beta1.Resource{
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

	cmd.Flags().StringArrayVarP(&values, "values", "v", []string{}, "list of key-value pair. Example: key=value foo=bar")
	cmd.MarkFlagRequired("values")
	cmd.PersistentFlags().UintVarP(&id, "id", "i", 0, "resource id")
	cmd.MarkPersistentFlagRequired("id")

	return cmd
}

func metadataGetCmd(c *app.CLIConfig) *cobra.Command {
	var format string
	var id uint

	cmd := &cobra.Command{
		Use:   "metadata-get",
		Short: "Get metadata for a resource",
		Example: heredoc.Doc(`
			$ guardian resource metadata-get --id=<resource-id>
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

			res, err := client.GetResource(ctx, &guardianv1beta1.GetResourceRequest{
				Id: uint32(id),
			})
			if err != nil {
				return err
			}

			resource := res.GetResource()
			details := resource.GetDetails()
			detailsMap := details.AsMap()
			if len(detailsMap) == 0 {
				fmt.Println("This resource has no metadata to show.")
				return nil
			}

			fmt.Print("DETAILS\n\n")
			for key, value := range detailsMap {
				fmt.Println(key, ":", value)

			}
			return nil
		},
	}

	cmd.Flags().UintVarP(&id, "id", "i", 0, "Resource id")
	cmd.MarkFlagRequired("id")
	cmd.Flags().StringVarP(&format, "format", "f", "", "Format of output - json yaml prettyjson etc")
	return cmd
}
