package grant_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/odpf/guardian/core/grant"
	"github.com/odpf/guardian/core/grant/mocks"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository *mocks.Repository
	service        *grant.Service
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) setup() {
	s.mockRepository = new(mocks.Repository)
	s.service = grant.NewService(grant.ServiceDeps{
		Repository: s.mockRepository,
		Logger:     log.NewNoop(),
	})
}

func (s *ServiceTestSuite) TestList() {
	s.Run("should return list of grant on success", func() {
		s.setup()

		filter := domain.ListGrantsFilter{}
		expectedGrants := []domain.Grant{}
		s.mockRepository.EXPECT().
			List(mock.AnythingOfType("*context.emptyCtx"), filter).
			Return(expectedGrants, nil).Once()

		grants, err := s.service.List(context.Background(), filter)

		s.NoError(err)
		s.Equal(expectedGrants, grants)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if repository returns an error", func() {
		s.setup()

		expectedError := errors.New("unexpected error")
		s.mockRepository.EXPECT().
			List(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("domain.ListGrantsFilter")).
			Return(nil, expectedError).Once()

		grants, err := s.service.List(context.Background(), domain.ListGrantsFilter{})

		s.ErrorIs(err, expectedError)
		s.Nil(grants)
		s.mockRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestGetByID() {
	s.Run("should return grant details on success", func() {
		s.setup()

		id := uuid.New().String()
		expectedGrant := &domain.Grant{}
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), id).
			Return(expectedGrant, nil).
			Once()

		grant, err := s.service.GetByID(context.Background(), id)

		s.NoError(err)
		s.Equal(expectedGrant, grant)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if id param is empty", func() {
		s.setup()

		expectedError := grant.ErrEmptyIDParam

		grant, err := s.service.GetByID(context.Background(), "")

		s.ErrorIs(err, expectedError)
		s.Nil(grant)
		s.mockRepository.AssertExpectations(s.T())
	})

	s.Run("should return error if repository returns an error", func() {
		s.setup()

		expectedError := errors.New("unexpected error")
		s.mockRepository.EXPECT().
			GetByID(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string")).
			Return(nil, expectedError).Once()

		grant, err := s.service.GetByID(context.Background(), "test-id")

		s.ErrorIs(err, expectedError)
		s.Nil(grant)
		s.mockRepository.AssertExpectations(s.T())
	})
}
