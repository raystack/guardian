package grant_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
		Validator:  validator.New(),
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

func (s *ServiceTestSuite) TestPrepare() {
	s.Run("should return error if appeal is invalid", func() {
		testCases := []struct {
			name   string
			appeal domain.Appeal
		}{
			{
				"appeal status is not approved",
				domain.Appeal{
					Status:      domain.AppealStatusPending,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "test-resource-id",
				},
			},
			{
				"account id is empty",
				domain.Appeal{
					Status:      domain.AppealStatusApproved,
					AccountID:   "",
					AccountType: "user",
					ResourceID:  "test-resource-id",
				},
			},
			{
				"account type is empty",
				domain.Appeal{
					Status:      domain.AppealStatusApproved,
					AccountID:   "user@example.com",
					AccountType: "",
					ResourceID:  "test-resource-id",
				},
			},
			{
				"resource id is empty",
				domain.Appeal{
					Status:      domain.AppealStatusApproved,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "",
				},
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.setup()
				actualGrant, actualError := s.service.Prepare(context.Background(), tc.appeal)

				s.Error(actualError)
				s.Nil(actualGrant)
			})
		}
	})

	s.Run("should return valid grant", func() {
		expDate := time.Now().Add(24 * time.Hour)
		testCases := []struct {
			name          string
			appeal        domain.Appeal
			expectedGrant *domain.Grant
		}{
			{
				name: "appeal with empty permanent duration option",
				appeal: domain.Appeal{
					ID:          "test-appeal-id",
					Status:      domain.AppealStatusApproved,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "test-user-id",
					Role:        "test-role",
					Permissions: []string{"test-permissions"},
					CreatedBy:   "user@example.com",
				},
				expectedGrant: &domain.Grant{
					Status:      domain.GrantStatusActive,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "test-user-id",
					Role:        "test-role",
					Permissions: []string{"test-permissions"},
					AppealID:    "test-appeal-id",
					CreatedBy:   "user@example.com",
					IsPermanent: true,
				},
			}, {
				name: "appeal with 0h as permanent duration option",
				appeal: domain.Appeal{
					ID:          "test-appeal-id",
					Status:      domain.AppealStatusApproved,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "test-user-id",
					Role:        "test-role",
					Permissions: []string{"test-permissions"},
					CreatedBy:   "user@example.com",
					Options: &domain.AppealOptions{
						Duration: "0h",
					},
				},
				expectedGrant: &domain.Grant{
					Status:      domain.GrantStatusActive,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "test-user-id",
					Role:        "test-role",
					Permissions: []string{"test-permissions"},
					AppealID:    "test-appeal-id",
					CreatedBy:   "user@example.com",
					IsPermanent: true,
				},
			},
			{
				name: "appeal with duration option",
				appeal: domain.Appeal{
					ID:          "test-appeal-id",
					Status:      domain.AppealStatusApproved,
					AccountID:   "user@example.com",
					AccountType: "user",
					ResourceID:  "test-user-id",
					Role:        "test-role",
					Permissions: []string{"test-permissions"},
					CreatedBy:   "user@example.com",
					Options: &domain.AppealOptions{
						Duration: "24h",
					},
				},
				expectedGrant: &domain.Grant{
					Status:         domain.GrantStatusActive,
					AccountID:      "user@example.com",
					AccountType:    "user",
					ResourceID:     "test-user-id",
					Role:           "test-role",
					Permissions:    []string{"test-permissions"},
					AppealID:       "test-appeal-id",
					CreatedBy:      "user@example.com",
					IsPermanent:    false,
					ExpirationDate: &expDate,
				},
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.setup()
				actualGrant, actualError := s.service.Prepare(context.Background(), tc.appeal)

				s.NoError(actualError)
				if diff := cmp.Diff(tc.expectedGrant, actualGrant, cmpopts.EquateApproxTime(time.Second)); diff != "" {
					s.T().Errorf("result not match, diff: %v", diff)
				}
			})
		}
	})
}
