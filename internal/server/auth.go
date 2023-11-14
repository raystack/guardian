package server

import (
	"context"

	"github.com/raystack/guardian/pkg/auth"

	ctx_logrus "github.com/grpc-ecosystem/go-grpc-middleware/tags/logrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var logrusActorKey = "actor"

func withAuthenticatedUserEmail(headerKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if v := md.Get(headerKey); len(v) > 0 {
				userEmail := v[0]
				if len(userEmail) > 0 {
					ctx = auth.WrapEmailInCtx(ctx, userEmail)
				}
			}
		}

		return handler(ctx, req)
	}
}

func withLogrusContext() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if userEmail := auth.FetchEmailFromCtx(ctx); userEmail != "" {
			ctx_logrus.AddFields(ctx, logrus.Fields{
				logrusActorKey: userEmail,
			})
		}

		return handler(ctx, req)
	}
}
