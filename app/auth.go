package app

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type authenticatedUserEmailContextKey struct{}

func withAuthenticatedUserEmail(headerKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if v := md.Get(headerKey); len(v) > 0 {
				actor := v[0]
				ctx = context.WithValue(ctx, authenticatedUserEmailContextKey{}, actor)
			}
		}

		return handler(ctx, req)
	}
}
