package app

import (
	"context"

	ctx_logrus "github.com/grpc-ecosystem/go-grpc-middleware/tags/logrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type authenticatedUserEmailContextKey struct{}

func withAuthenticatedUserEmail(headerKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if v := md.Get(headerKey); len(v) > 0 {
				userEmail := v[0]
				ctx = context.WithValue(ctx, authenticatedUserEmailContextKey{}, userEmail)

				ctx_logrus.AddFields(ctx, logrus.Fields{
					headerKey: userEmail,
				})
			}
		}

		return handler(ctx, req)
	}
}
