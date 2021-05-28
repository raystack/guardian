package iam_test

import (
	"testing"

	"github.com/odpf/guardian/iam"
	"github.com/odpf/guardian/mocks"
	"github.com/stretchr/testify/mock"
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

func (s *ServiceTestSuite) TestGetUserApproverEmails() {
	s.Run("should return error if email param is empty", func() {
		actualResult, actualError := s.service.GetUserApproverEmails("")

		s.Nil(actualResult)
		s.EqualError(actualError, iam.ErrEmptyUserEmailParam.Error())
	})

	s.Run("should pass user as the query string to client", func() {
		user := "test@email.com"
		expectedQuery := map[string]string{
			"user": user,
		}
		s.mockClient.On("GetUserApproverEmails", expectedQuery).Return(nil, nil).Once()

		s.service.GetUserApproverEmails(user)

		s.mockClient.AssertExpectations(s.T())
	})

	s.Run("should return error if approver emails are empty", func() {
		user := "test@email.com"
		s.mockClient.On("GetUserApproverEmails", mock.Anything).Return([]string{}, nil).Once()

		actualResult, actualError := s.service.GetUserApproverEmails(user)

		s.Nil(actualResult)
		s.EqualError(actualError, iam.ErrEmptyApprovers.Error())
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
