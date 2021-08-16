package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "github.com/odpf/guardian/api/handler/v1"
	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/odpf/guardian/domain"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func policiesCommand(c *config, adapter v1.ProtoAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policies",
		Short: "manage policies",
	}

	cmd.AddCommand(listPoliciesCommand(c))
	cmd.AddCommand(createPolicyCommand(c, adapter))
	cmd.AddCommand(updatePolicyCommand(c, adapter))

	return cmd
}

func listPoliciesCommand(c *config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list policies",
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
			res, err := client.ListPolicies(requestTimeoutCtx, &pb.ListPoliciesRequest{})
			if err != nil {
				return err
			}

			t := getTablePrinter(os.Stdout, []string{"ID", "VERSION", "DESCRIPTION", "STEPS"})
			for _, p := range res.GetPolicies() {
				var stepNames []string
				for _, s := range p.GetSteps() {
					stepNames = append(stepNames, s.GetName())
				}
				t.Append([]string{
					fmt.Sprintf("%v", p.GetId()),
					fmt.Sprintf("%v", p.GetVersion()),
					p.GetDescription(),
					strings.Join(stepNames, ","),
				})
			}
			t.Render()
			return nil
		},
	}
}

func createPolicyCommand(c *config, adapter v1.ProtoAdapter) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			var policy domain.Policy
			switch filepath.Ext(filePath) {
			case ".json":
				if err := json.Unmarshal(b, &policy); err != nil {
					return err
				}
			case ".yaml", ".yml":
				if err := yaml.Unmarshal(b, &policy); err != nil {
					return err
				}
			default:
				return errors.New("unsupported file type")
			}
			policyProto, err := adapter.ToPolicyProto(&policy)
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
			res, err := client.CreatePolicy(requestTimeoutCtx, &pb.CreatePolicyRequest{
				Policy: policyProto,
			})
			if err != nil {
				return err
			}

			fmt.Printf("policy created with id: %v", res.GetPolicy().GetId())

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "path to the policy config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func updatePolicyCommand(c *config, adapter v1.ProtoAdapter) *cobra.Command {
	var id string
	var filePath string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			var policy domain.Policy
			switch filepath.Ext(filePath) {
			case ".json":
				if err := json.Unmarshal(b, &policy); err != nil {
					return err
				}
			case ".yaml", ".yml":
				if err := yaml.Unmarshal(b, &policy); err != nil {
					return err
				}
			default:
				return errors.New("unsupported file type")
			}
			policyProto, err := adapter.ToPolicyProto(&policy)
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
			policyID := id
			if policyID == "" {
				policyID = policyProto.GetId()
			}
			_, err = client.UpdatePolicy(requestTimeoutCtx, &pb.UpdatePolicyRequest{
				Id:     policyID,
				Policy: policyProto,
			})
			if err != nil {
				return err
			}

			fmt.Println("policy updated")

			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "policy id")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "path to the policy config")
	cmd.MarkFlagRequired("file")

	return cmd
}
