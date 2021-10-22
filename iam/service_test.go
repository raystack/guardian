package iam_test

import (
	"testing"

	"github.com/odpf/guardian/iam"
	"github.com/odpf/guardian/mocks"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockClient *mocks.IAMClient
	service    *iam.Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockClient = new(mocks.IAMClient)
	s.service = iam.NewService(s.mockClient)
}

func (s *ServiceTestSuite) TestGetUser() {
	s.Run("should return error if id param is empty", func() {
		actualResult, actualError := s.service.GetUser("")

		s.Nil(actualResult)
		s.EqualError(actualError, iam.ErrEmptyUserEmailParam.Error())
	})

	s.Run("should pass user id as to the client", func() {
		user := "test@email.com"
		s.mockClient.On("GetUser", user).Return(nil, nil).Once()

		s.service.GetUser(user)

		s.mockClient.AssertExpectations(s.T())
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
