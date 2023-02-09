package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/odpf/guardian/pkg/auth"
	"github.com/odpf/guardian/pkg/auth/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var authContextValues = map[string]string{
	"Authorization": "Bearer some-bearer-token-in-JWT",
}

type InterceptorTestSuite struct {
	suite.Suite
}

func (s *InterceptorTestSuite) TestIdTokenValidator_WithBearerTokenValidator() {
	emptyAuthContextValues := map[string]string{
		"Authorization": "Bearer  ",
	}

	testCases := []struct {
		name        string
		params      auth.OIDCValidatorParams
		ctx         context.Context
		mockFunc    func(validator *mocks.OIDCValidator)
		expectedErr error
	}{
		{
			name:        "MD context value does not exist",
			params:      auth.OIDCValidatorParams{},
			ctx:         context.Background(),
			mockFunc:    func(validator *mocks.OIDCValidator) {},
			expectedErr: auth.InvalidAuthError,
		},
		{
			name:        "empty authorization header",
			params:      auth.OIDCValidatorParams{},
			ctx:         metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{})),
			mockFunc:    func(validator *mocks.OIDCValidator) {},
			expectedErr: auth.InvalidAuthError,
		},
		{
			name:        "empty bearer token on authorization header",
			params:      auth.OIDCValidatorParams{},
			ctx:         metadata.NewIncomingContext(context.Background(), metadata.New(emptyAuthContextValues)),
			mockFunc:    func(validator *mocks.OIDCValidator) {},
			expectedErr: auth.InvalidAuthError,
		},
		{
			name: "error while validating token",
			params: auth.OIDCValidatorParams{
				Audience: "google.com",
			},
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(authContextValues)),
			mockFunc: func(validator *mocks.OIDCValidator) {
				validator.On("Validate", mock.Anything, mock.Anything, "google.com").
					Return(nil, errors.New("something happened"))
			},
			expectedErr: auth.InvalidAuthError,
		},
		{
			name: "email domain does not match with eligible domains",
			params: auth.OIDCValidatorParams{
				Audience:          "google.com",
				ValidEmailDomains: "example.com,something.org",
			},
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(authContextValues)),
			mockFunc: func(validator *mocks.OIDCValidator) {

				payload := &idtoken.Payload{
					Claims: map[string]interface{}{
						"email": "something@gmail.com",
					},
				}
				validator.On("Validate", mock.Anything, mock.Anything, "google.com").
					Return(payload, nil)
			},
			expectedErr: auth.InvalidAuthError,
		},
		{
			name: "successful request with matching eligible email domains",
			params: auth.OIDCValidatorParams{
				Audience:          "google.com",
				ValidEmailDomains: "example.com,something.org",
			},
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(authContextValues)),
			mockFunc: func(validator *mocks.OIDCValidator) {
				payload := &idtoken.Payload{
					Claims: map[string]interface{}{
						"email": "something@example.com",
					},
				}
				validator.On("Validate", mock.Anything, mock.Anything, "google.com").
					Return(payload, nil)
			},
			expectedErr: nil,
		},
		{
			name: "successful request with no eligible email domains configurations whatsoever",
			params: auth.OIDCValidatorParams{
				Audience: "google.com",
			},
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(authContextValues)),
			mockFunc: func(validator *mocks.OIDCValidator) {
				payload := &idtoken.Payload{
					Claims: map[string]interface{}{
						"email": "something@example.com",
					},
				}
				validator.On("Validate", mock.Anything, mock.Anything, "google.com").
					Return(payload, nil)
			},
			expectedErr: nil,
		},
	}

	var req interface{}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			validator := new(mocks.OIDCValidator)
			authValidator := auth.NewOIDCValidator(validator, tc.params)
			interceptFunc := authValidator.WithOIDCValidator()

			tc.mockFunc(validator)
			result, err := interceptFunc(tc.ctx, req, &grpc.UnaryServerInfo{}, s.unaryDummyHandler)

			assert.Nil(s.T(), result)
			assert.Equal(s.T(), tc.expectedErr, err)
		})
	}
}

func (suite *InterceptorTestSuite) unaryDummyHandler(ctx context.Context, _ interface{}) (interface{}, error) {
	expectedCtx := metadata.NewIncomingContext(context.Background(), metadata.New(authContextValues))
	expectedCtx = context.WithValue(expectedCtx, auth.OIDCEmailContextKey{}, "something@example.com")

	assert.Equal(suite.T(), expectedCtx, ctx, "final method handler doesn't have matching context")

	return nil, nil
}

func TestOidcValidatorInterceptor(t *testing.T) {
	suite.Run(t, new(InterceptorTestSuite))
}
