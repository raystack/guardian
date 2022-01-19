package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	handlerv1beta1 "github.com/odpf/guardian/api/handler/v1beta1"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
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
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
	}

	cmd.AddCommand(listResourcesCmd(c, adapter))
	cmd.AddCommand(getResourceCmd(c, adapter))
	cmd.AddCommand(metadataCmd(c))
	cmd.PersistentFlags().StringP("output", "o", "", "Print output with specified format (yaml,json,prettyjson)")

	return cmd
}

func listResourcesCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var providerType, providerURN, resourceType, resourceURN, name string
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
			s := term.Spin("Fetching resource list")
			defer s.Stop()

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

			format := cmd.Flag("output").Value.String()
			if format != "" {
				var resources []*domain.Resource
				for _, r := range res.GetResources() {
					resources = append(resources, adapter.FromResourceProto(r))
				}

				s.Stop()

				if err := printer.Text(resources, format); err != nil {
					return fmt.Errorf("failed to parse resources: %v", err)
				}
				return nil
			}

			report := [][]string{}

			resources := res.GetResources()

			s.Stop()

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

	cmd.Flags().StringVar(&providerType, "provider-type", "", "Filter by provider type")
	cmd.Flags().StringVar(&providerURN, "provider-urn", "", "Filter by provider urn")
	cmd.Flags().StringVar(&resourceType, "type", "", "Filter by type")
	cmd.Flags().StringVar(&resourceURN, "urn", "", "Filter by urn")
	cmd.Flags().StringVar(&name, "name", "", "Filter by name")
	cmd.Flags().StringArrayVar(&details, "details", nil, "Filter by details object values. Example: --details=key1.key2:value")
	cmd.Flags().BoolVar(&isDeleted, "deleted", false, "Show deleted resources")

	return cmd
}

func getResourceCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	return &cobra.Command{
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
			s := term.Spin("Fetching resource")
			defer s.Stop()

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			id := args[0]
			res, err := client.GetResource(ctx, &guardianv1beta1.GetResourceRequest{
				Id: id,
			})
			if err != nil {
				return err
			}

			r := adapter.FromResourceProto(res.GetResource())

			s.Stop()

			format := cmd.Flag("output").Value.String()
			if format == "" {
				format = "yaml"
			}
			if err := printer.Text(r, format); err != nil {
				return fmt.Errorf("failed to parse resource: %v", err)
			}
			return nil
		},
	}
}

func metadataCmd(c *app.CLIConfig) *cobra.Command {
	var id string
	var values []string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Store new metadata for a resource",
		Example: heredoc.Doc(`
			$ guardian resource metadata set values foo=bar
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			s := term.Spin("Updating resource")
			defer s.Stop()

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
				Id: id,
				Resource: &guardianv1beta1.Resource{
					Details: metadataProto,
				},
			})
			if err != nil {
				return err
			}

			s.Stop()

			fmt.Println("Successfully updated metadata")

			return nil
		},
	}

	cmd.Flags().StringArrayVar(&values, "values", []string{}, "list of key-value pair. Example: key=value foo=bar")

	cmd.PersistentFlags().StringVar(&id, "id", "", "resource id")
	cmd.MarkPersistentFlagRequired("id")

	return cmd
}
