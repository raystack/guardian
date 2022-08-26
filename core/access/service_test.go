package access_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/odpf/guardian/core/access"
	"github.com/odpf/guardian/core/access/mocks"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository *mocks.Repository
	service        *access.Service
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) setup() {
	s.mockRepository = new(mocks.Repository)
	s.service = access.NewService(access.ServiceDeps{
		Repository: s.mockRepository,
		Logger:     log.NewNoop(),
	})
}

func (s *ServiceTestSuite) TestList() {
	s.Run("should return list of access on success", func() {
		s.setup()

		filter := domain.ListAccessesFilter{}
		expectedAccesses := []domain.Access{}
		s.mockRepository.EXPECT().
			List(mock.AnythingOfType("*context.emptyCtx"), filter).
			Return(expectedAccesses, nil).Once()

		accesses, err := s.service.List(context.Background(), filter)

		s.NoError(err)
		s.Equal(expectedAccesses, accesses)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if repository returns an error", func() {
		s.setup()

		expectedError := errors.New("unexpected error")
		s.mockRepository.EXPECT().
			List(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("domain.ListAccessesFilter")).
			Return(nil, expectedError).Once()

		accesses, err := s.service.List(context.Background(), domain.ListAccessesFilter{})

		s.ErrorIs(err, expectedError)
		s.Nil(accesses)
		s.mockRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestGetByID() {
	s.Run("should return access details on success", func() {
		s.setup()

		id := uuid.New().String()
		expectedAccess := &domain.Access{}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), id).
			Return(expectedAccess, nil).
			Once()

		access, err := s.service.GetByID(context.Background(), id)

		s.NoError(err)
		s.Equal(expectedAccess, access)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if id param is empty", func() {
		s.setup()

		expectedError := access.ErrEmptyIDParam

		access, err := s.service.GetByID(context.Background(), "")

		s.ErrorIs(err, expectedError)
		s.Nil(access)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if repository returns an error", func() {
		s.setup()

		expectedError := errors.New("unexpected error")
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string")).
			Return(nil, expectedError).Once()

		access, err := s.service.GetByID(context.Background(), "test-id")

		s.ErrorIs(err, expectedError)
		s.Nil(access)
		s.mockRepository.AssertExpectations(s.T())
	})
}
