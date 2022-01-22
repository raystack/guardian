package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	handlerv1beta1 "github.com/odpf/guardian/api/handler/v1beta1"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"
)

func ResourceCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource",
		Aliases: []string{"resources"},
		Short:   "Manage resources",
		Example: heredoc.Doc(`
			$ guardian resource list
			$ guardian resource view
			$ guardian resource set
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
	}

	cmd.AddCommand(listResourcesCmd(c, adapter))
	cmd.AddCommand(viewResourceCmd(c, adapter))
	cmd.AddCommand(setResourceCmd(c, adapter))
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

func viewResourceCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var format string
	var metadata bool

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View a resource details",
		Example: heredoc.Doc(`
			$ guardian resource view <resource-id> --format=json --metadata=true
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
			_, err = strconv.ParseUint(id, 10, 32)
			if err != nil {
				return fmt.Errorf("invalid provider id: %v", err)
			}

			res, err := client.GetResource(ctx, &guardianv1beta1.GetResourceRequest{
				Id: id,
			})
			if err != nil {
				return err
			}

			if format != "" {
				r := adapter.FromResourceProto(res.GetResource())
				if err := printer.Text(r, format); err != nil {
					return fmt.Errorf("failed to parse resources: %v", err)
				}
			} else {

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
			}

			if metadata {
				r := res.GetResource()
				d := r.GetDetails()
				details := d.AsMap()
				if len(details) == 0 {
					fmt.Println("This resource has no metadata to show.")
					return nil
				}

				fmt.Print("\nMETADATA\n")
				for key, value := range details {
					fmt.Println(key, ":", value)

				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&metadata, "metadata", "m", false, "Set if you want to see metadata")
	cmd.Flags().StringVarP(&format, "format", "f", "", "Format of output - json yaml prettyjson etc")
	return cmd
}

func setResourceCmd(c *app.CLIConfig, adapter handlerv1beta1.ProtoAdapter) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Store new metadata for a resource",
		Example: heredoc.Doc(`
			$ guardian resource set <resource-id> --filePath=<file-path>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := term.Spin("Updating resource")
			defer s.Stop()

			var resource domain.Resource
			if err := parseFile(filePath, &resource); err != nil {
				return err
			}

			resourceProto, err := adapter.ToResourceProto(&resource)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, cancel, err := createClient(ctx, c.Host)
			if err != nil {
				return err
			}
			defer cancel()

			id := args[0]
			_, err = strconv.ParseUint(id, 10, 32)
			if err != nil {
				return fmt.Errorf("invalid provider id: %v", err)
			}

			_, err = client.UpdateResource(ctx, &guardianv1beta1.UpdateResourceRequest{
				Id:       id,
				Resource: resourceProto,
			})
			if err != nil {
				return err
			}

			s.Stop()

			fmt.Println("Successfully updated metadata")

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "updated resource file path")
	cmd.MarkFlagRequired("file")

	return cmd
}
