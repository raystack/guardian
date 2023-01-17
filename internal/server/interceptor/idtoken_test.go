package interceptor_test

import (
	"context"
	"errors"
	"github.com/odpf/guardian/internal/server"
	"github.com/odpf/guardian/internal/server/interceptor"
	"github.com/odpf/guardian/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"testing"
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
		params      *interceptor.IdTokenValidatorParams
		ctx         context.Context
		mockFunc    func(validator *mocks.IdTokenValidator)
		expectedErr error
	}{
		{
			name:        "MD context value does not exist",
			params:      &interceptor.IdTokenValidatorParams{},
			ctx:         context.Background(),
			mockFunc:    func(validator *mocks.IdTokenValidator) {},
			expectedErr: interceptor.InvalidAuthError,
		},
		{
			name:        "empty authorization header",
			params:      &interceptor.IdTokenValidatorParams{},
			ctx:         metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{})),
			mockFunc:    func(validator *mocks.IdTokenValidator) {},
			expectedErr: interceptor.InvalidAuthError,
		},
		{
			name:        "empty bearer token on authorization header",
			params:      &interceptor.IdTokenValidatorParams{},
			ctx:         metadata.NewIncomingContext(context.Background(), metadata.New(emptyAuthContextValues)),
			mockFunc:    func(validator *mocks.IdTokenValidator) {},
			expectedErr: interceptor.InvalidAuthError,
		},
		{
			name: "error while validating token",
			params: &interceptor.IdTokenValidatorParams{
				Audience: "google.com",
			},
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(authContextValues)),
			mockFunc: func(validator *mocks.IdTokenValidator) {
				validator.On("Validate", mock.Anything, mock.Anything, "google.com").
					Return(nil, errors.New("something happened"))
			},
			expectedErr: interceptor.InvalidAuthError,
		},
		{
			name: "email domain does not match with eligible domains",
			params: &interceptor.IdTokenValidatorParams{
				Audience:          "google.com",
				ValidEmailDomains: "example.com,something.org",
			},
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(authContextValues)),
			mockFunc: func(validator *mocks.IdTokenValidator) {

				payload := &idtoken.Payload{
					Claims: map[string]interface{}{
						"email": "something@gmail.com",
					},
				}
				validator.On("Validate", mock.Anything, mock.Anything, "google.com").
					Return(payload, nil)
			},
			expectedErr: interceptor.InvalidAuthError,
		},
		{
			name: "successful request with matching eligible email domains",
			params: &interceptor.IdTokenValidatorParams{
				Audience:          "google.com",
				ValidEmailDomains: "example.com,something.org",
				ContextKey:        server.AuthenticatedUserEmailContextKey{},
			},
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(authContextValues)),
			mockFunc: func(validator *mocks.IdTokenValidator) {
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
			params: &interceptor.IdTokenValidatorParams{
				Audience:   "google.com",
				ContextKey: server.AuthenticatedUserEmailContextKey{},
			},
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(authContextValues)),
			mockFunc: func(validator *mocks.IdTokenValidator) {
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
			validator := new(mocks.IdTokenValidator)
			authValidator := interceptor.NewIdTokenValidator(validator, tc.params)
			interceptFunc := authValidator.WithBearerTokenValidator()

			tc.mockFunc(validator)
			result, err := interceptFunc(tc.ctx, req, &grpc.UnaryServerInfo{}, s.unaryDummyHandler)

			assert.Nil(s.T(), result)
			assert.Equal(s.T(), tc.expectedErr, err)
		})
	}
}

func (suite *InterceptorTestSuite) unaryDummyHandler(ctx context.Context, _ interface{}) (interface{}, error) {
	expectedCtx := metadata.NewIncomingContext(context.Background(), metadata.New(authContextValues))
	expectedCtx = context.WithValue(expectedCtx, server.AuthenticatedUserEmailContextKey{}, "something@example.com")

	assert.Equal(suite.T(), expectedCtx, ctx, "final method handler doesn't have matching context")

	return nil, nil
}

func TestIdTokenValidatorInterceptor(t *testing.T) {
	suite.Run(t, new(InterceptorTestSuite))
}
