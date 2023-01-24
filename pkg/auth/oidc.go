package auth

import (
	"context"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var InvalidAuthError = status.Errorf(codes.Unauthenticated, "invalid authentication credentials")

type Validator interface {
	Validate(ctx context.Context, token string, audience string) (*idtoken.Payload, error)
}

type OidcValidator struct {
	validator         Validator
	audience          string
	validEmailDomains []string
	headerKey         string
	contextKey        interface{}
}

type OidcValidatorParams struct {
	Audience          string
	ValidEmailDomains string
	HeaderKey         string
	ContextKey        interface{}
}

func NewOidcValidator(validator Validator, config *OidcValidatorParams) *OidcValidator {
	audience := config.Audience
	headerKey := config.HeaderKey

	var validEmailDomains []string
	if strings.TrimSpace(config.ValidEmailDomains) != "" {
		validEmailDomains = strings.Split(config.ValidEmailDomains, ",")
	}

	return &OidcValidator{
		validator:         validator,
		audience:          audience,
		validEmailDomains: validEmailDomains,
		headerKey:         headerKey,
		contextKey:        config.ContextKey,
	}
}

func (v *OidcValidator) WithOidcValidator() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, InvalidAuthError
		}

		headerValue := md.Get("authorization")
		if len(headerValue) == 0 || strings.TrimSpace(headerValue[0]) == "" {
			return nil, InvalidAuthError
		}

		bearerToken := strings.TrimSpace(strings.TrimPrefix(headerValue[0], "Bearer "))
		if len(bearerToken) == 0 {
			return nil, InvalidAuthError
		}

		payload, err := v.validator.Validate(ctx, bearerToken, v.audience)
		if err != nil {
			return nil, InvalidAuthError
		}

		email := payload.Claims["email"].(string)
		if err := v.validateEmailDomain(email); err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, v.contextKey, email)
		ctxlogrus.AddFields(ctx, logrus.Fields{
			v.headerKey: email,
		})

		return handler(ctx, req)
	}
}

func (v *OidcValidator) validateEmailDomain(email string) error {
	// no valid email domains listed means that no email domain will be checked
	if len(v.validEmailDomains) == 0 {
		return nil
	}

	emailDomainMatch := false
	for _, validEmailDomain := range v.validEmailDomains {
		if strings.HasSuffix(email, "@"+validEmailDomain) {
			emailDomainMatch = true
			break
		}
	}

	if !emailDomainMatch {
		return InvalidAuthError
	}
	return nil
}
