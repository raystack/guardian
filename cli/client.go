package cli

import (
	"context"
	"errors"
	"time"

	guardianv1beta1 "github.com/goto/guardian/api/proto/gotocompany/guardian/v1beta1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

func createConnection(ctx context.Context, host string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}

	return grpc.DialContext(ctx, host, opts...)
}

func createClient(cmd *cobra.Command) (guardianv1beta1.GuardianServiceClient, func(), error) {
	host, err := cmd.Flags().GetString("host")
	if err != nil {
		return nil, nil, err
	}
	if host == "" {
		return nil, nil, errors.New("\"host\" not set")
	}

	dialTimeoutCtx, dialCancel := context.WithTimeout(cmd.Context(), time.Second*2)
	conn, err := createConnection(dialTimeoutCtx, host)
	if err != nil {
		dialCancel()
		return nil, nil, err
	}

	cancel := func() {
		dialCancel()
		conn.Close()
	}

	client := guardianv1beta1.NewGuardianServiceClient(conn)
	return client, cancel, nil
}
