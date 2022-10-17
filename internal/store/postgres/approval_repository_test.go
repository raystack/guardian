package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type ApprovalRepositoryTestSuite struct {
	suite.Suite
	store      *postgres.Store
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.ApprovalRepository

	dummyProvider *domain.Provider
	dummyPolicy   *domain.Policy
	dummyResource *domain.Resource
	dummyAppeal   *domain.Appeal
}

func TestApprovalRepository(t *testing.T) {
	suite.Run(t, new(ApprovalRepositoryTestSuite))
}

func (s *ApprovalRepositoryTestSuite) SetupSuite() {
	var err error
	logger := log.NewLogrus(log.LogrusWithLevel("debug"))
	s.store, s.pool, s.resource, err = newTestStore(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.repository = postgres.NewApprovalRepository(s.store.DB())

	ctx := context.Background()

	s.dummyPolicy = &domain.Policy{
		ID:      "policy_test",
		Version: 1,
	}
	policyRepository := postgres.NewPolicyRepository(s.store.DB())
	err = policyRepository.Create(ctx, s.dummyPolicy)
	s.Require().NoError(err)

	s.dummyProvider = &domain.Provider{
		Type: "provider_test",
		URN:  "provider_urn_test",
		Config: &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type: "resource_type_test",
					Policy: &domain.PolicyConfig{
						ID:      s.dummyPolicy.ID,
						Version: int(s.dummyPolicy.Version),
					},
				},
			},
		},
	}
	providerRepository := postgres.NewProviderRepository(s.store.DB())
	err = providerRepository.Create(ctx, s.dummyProvider)
	s.Require().NoError(err)

	s.dummyResource = &domain.Resource{
		ProviderType: s.dummyProvider.Type,
		ProviderURN:  s.dummyProvider.URN,
		Type:         "resource_type_test",
		URN:          "resource_urn_test",
		Name:         "resource_name_test",
	}
	resourceRepository := postgres.NewResourceRepository(s.store.DB())
	err = resourceRepository.BulkUpsert(ctx, []*domain.Resource{s.dummyResource})
	s.Require().NoError(err)

	s.dummyAppeal = &domain.Appeal{
		ResourceID:    s.dummyResource.ID,
		PolicyID:      s.dummyPolicy.ID,
		PolicyVersion: s.dummyPolicy.Version,
		AccountID:     "user@example.com",
		AccountType:   domain.DefaultAppealAccountType,
		Role:          "role_test",
		Permissions:   []string{"permission_test"},
		CreatedBy:     "user@example.com",
	}
	appealRepository := postgres.NewAppealRepository(s.store.DB())
	err = appealRepository.BulkUpsert([]*domain.Appeal{s.dummyAppeal})
	s.Require().NoError(err)
}

func (s *ApprovalRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	db, err := s.store.DB().DB()
	if err != nil {
		s.T().Fatal(err)
	}
	err = db.Close()
	if err != nil {
		s.T().Fatal(err)
	}

	err = purgeTestDocker(s.pool, s.resource)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *ApprovalRepositoryTestSuite) TestListApprovals() {
	dummyApprovals := []*domain.Approval{
		{
			Name:          "test-approval-name-1",
			Index:         0,
			AppealID:      s.dummyAppeal.ID,
			Status:        "test-status-1",
			PolicyID:      "test-policy-id",
			PolicyVersion: 1,

			Appeal: s.dummyAppeal,
		},
		{
			Name:          "test-approval-name-2",
			Index:         1,
			AppealID:      s.dummyAppeal.ID,
			Status:        "test-status-2",
			PolicyID:      "test-policy-id",
			PolicyVersion: 1,
			Appeal:        s.dummyAppeal,
		},
	}

	err := s.repository.BulkInsert(dummyApprovals)
	s.Require().NoError(err)

	dummyApprover := []*domain.Approver{
		{
			ApprovalID: dummyApprovals[0].ID,
			AppealID:   s.dummyAppeal.ID,
			Email:      "approver1@email.com",
		},
		{
			ApprovalID: dummyApprovals[1].ID,
			AppealID:   s.dummyAppeal.ID,
			Email:      "approver2@email.com",
		},
	}

	err = s.repository.AddApprover(dummyApprover[0])
	s.Require().NoError(err)
	err = s.repository.AddApprover(dummyApprover[1])
	s.Require().NoError(err)

	s.Run("should return list of approvals on success", func() {
		approvals, err := s.repository.ListApprovals(&domain.ListApprovalsFilter{
			AccountID: s.dummyAppeal.AccountID,
			CreatedBy: dummyApprover[0].Email,
			Statuses:  []string{"test-status-1"},
			OrderBy:   []string{"status", "updated_at:desc", "created_at"},
		})

		s.NoError(err)
		s.Len(approvals, 1)
		s.Equal(dummyApprovals[0].ID, approvals[0].ID)
	})

	s.Run("should return error if conditions invalid", func() {
		approvals, err := s.repository.ListApprovals(&domain.ListApprovalsFilter{
			AccountID: "",
			CreatedBy: "",
			Statuses:  []string{},
			OrderBy:   []string{},
		})

		s.Error(err)
		s.Nil(approvals)
	})

	s.Run("should return error if db execution returns an error on listing approvers", func() {
		approvals, err := s.repository.ListApprovals(&domain.ListApprovalsFilter{
			OrderBy: []string{"not-a-column"},
		})

		s.Error(err)
		s.Nil(approvals)
	})
}

func (s *ApprovalRepositoryTestSuite) TestBulkInsert() {
	actor := "user@email.com"
	approvals := []*domain.Approval{
		{
			Name:          "approval_step_1",
			Index:         0,
			AppealID:      s.dummyAppeal.ID,
			Status:        domain.ApprovalStatusPending,
			Actor:         &actor,
			PolicyID:      "policy_1",
			PolicyVersion: 1,
		},
		{
			Name:          "approval_step_2",
			Index:         1,
			AppealID:      s.dummyAppeal.ID,
			Status:        domain.ApprovalStatusPending,
			Actor:         &actor,
			PolicyID:      "policy_1",
			PolicyVersion: 1,
		},
	}

	s.Run("should return error if got any from transaction", func() {
		expectedError := errors.New("empty slice found")
		actualError := s.repository.BulkInsert([]*domain.Approval{})
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return nil error on success", func() {
		err := s.repository.BulkInsert(approvals)
		s.Nil(err)
	})
}

func (s *ApprovalRepositoryTestSuite) TestAddApprover() {
	actor := "user@email.com"
	dummyApprovals := []*domain.Approval{
		{
			Name:          "approval_step_1",
			Index:         0,
			AppealID:      s.dummyAppeal.ID,
			Status:        domain.ApprovalStatusPending,
			Actor:         &actor,
			PolicyID:      "policy_1",
			PolicyVersion: 1,
		},
		{
			Name:          "approval_step_2",
			Index:         1,
			AppealID:      s.dummyAppeal.ID,
			Status:        domain.ApprovalStatusPending,
			Actor:         &actor,
			PolicyID:      "policy_1",
			PolicyVersion: 1,
		},
	}

	err := s.repository.BulkInsert(dummyApprovals)
	s.Require().NoError(err)

	s.Run("should return nil error on success", func() {
		dummyApprover := &domain.Approver{
			ApprovalID: dummyApprovals[0].ID,
			Email:      "user@example.com",
			AppealID:   s.dummyAppeal.ID,
		}

		err := s.repository.AddApprover(dummyApprover)
		s.NoError(err)
	})

	s.Run("should return error if approver payload is invalid", func() {
		invalidApprover := &domain.Approver{
			ID: "invalid-uuid",
		}

		err := s.repository.AddApprover(invalidApprover)

		s.EqualError(err, "parsing approver: parsing uuid: invalid UUID length: 12")
	})
}

func (s *ApprovalRepositoryTestSuite) TestDeleteApprover() {
	actor := "user@email.com"
	dummyApprovals := []*domain.Approval{
		{
			Name:          "approval_step_1",
			Index:         0,
			AppealID:      s.dummyAppeal.ID,
			Status:        domain.ApprovalStatusPending,
			Actor:         &actor,
			PolicyID:      "policy_1",
			PolicyVersion: 1,
		},
		{
			Name:          "approval_step_2",
			Index:         1,
			AppealID:      s.dummyAppeal.ID,
			Status:        domain.ApprovalStatusPending,
			Actor:         &actor,
			PolicyID:      "policy_1",
			PolicyVersion: 1,
		},
	}

	err := s.repository.BulkInsert(dummyApprovals)
	s.Require().NoError(err)

	dummyApprover := &domain.Approver{
		ApprovalID: dummyApprovals[0].ID,
		Email:      "user@example.com",
		AppealID:   s.dummyAppeal.ID,
	}

	err = s.repository.AddApprover(dummyApprover)
	s.NoError(err)

	s.Run("should return nil error on success", func() {
		err := s.repository.DeleteApprover(dummyApprovals[0].ID, dummyApprover.Email)
		s.NoError(err)
	})

	s.Run("should return error if db returns an error", func() {
		err := s.repository.DeleteApprover("", "")
		s.Error(err)
	})
}
