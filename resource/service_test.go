package resource_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository *mocks.ResourceRepository
	service        *resource.Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(mocks.ResourceRepository)
	s.service = resource.NewService(s.mockRepository)
}

func (s *ServiceTestSuite) TestBulkUpsert() {
	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockRepository.On("BulkUpsert", mock.Anything).Return(expectedError).Once()

		actualError := s.service.BulkUpsert([]*domain.Resource{})

		s.EqualError(actualError, expectedError.Error())
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
