package auth

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type FrontierConfig struct {
	Host               string `mapstructure:"host"`
	NamespaceClaimsKey string `mapstructure:"namespace_claims_key" default:"project_id"`
}

type namespaceContextKey struct{}

// FrontierJWTInterceptor extracts the frontier jwt from the request metadata and
// set the context with the extracted jwt claims
func FrontierJWTInterceptor(conf FrontierConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if v := md.Get("authorization"); len(v) > 0 && len(v[0]) > 0 {
				var frontierToken string
				if strings.HasPrefix(v[0], "Bearer ") {
					frontierToken = strings.TrimPrefix(v[0], "Bearer ")
				}
				if frontierToken != "" {
					// TODO(kushsharma): we should validate the token using frontier public key
					insecureToken, err := jwt.ParseInsecure([]byte(frontierToken))
					if err == nil {
						if namespace, claimExists := insecureToken.Get(conf.NamespaceClaimsKey); claimExists {
							if id, err := uuid.Parse(namespace.(string)); err == nil {
								ctx = context.WithValue(ctx, namespaceContextKey{}, id)
							}
						}
					}
				}
			}
		}

		return handler(ctx, req)
	}
}

func FetchNamespace(ctx context.Context) uuid.UUID {
	if namespace, ok := ctx.Value(namespaceContextKey{}).(uuid.UUID); ok {
		return namespace
	}
	return uuid.Nil
}
