package auth

import (
	"context"
	"strings"

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

type OIDCEmailContextKey struct{}

type OIDCAuth struct {
	Audience             string `mapstructure:"audience"`
	EligibleEmailDomains string `mapstructure:"eligible_email_domains"`
}

type OIDCValidator struct {
	validator         Validator
	audience          string
	validEmailDomains []string
}

func NewOIDCValidator(validator Validator, config OIDCAuth) *OIDCValidator {
	audience := config.Audience

	var validEmailDomains []string
	if strings.TrimSpace(config.EligibleEmailDomains) != "" {
		validEmailDomains = strings.Split(config.EligibleEmailDomains, ",")
	}

	return &OIDCValidator{
		validator:         validator,
		audience:          audience,
		validEmailDomains: validEmailDomains,
	}
}

func (v *OIDCValidator) WithOIDCValidator() grpc.UnaryServerInterceptor {
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

		ctx = context.WithValue(ctx, OIDCEmailContextKey{}, email)

		return handler(ctx, req)
	}
}

func (v *OIDCValidator) validateEmailDomain(email string) error {
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
