package audit

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryServerInterceptor(authenticatedUserHeaderKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if v := md.Get(authenticatedUserHeaderKey); len(v) > 0 {
				ctx = WithActor(ctx, v[0])
			}
		}

		return handler(ctx, req)
	}
}
