package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/internal/store/postgres"
	"github.com/goto/guardian/pkg/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type ApprovalRepositoryTestSuite struct {
	suite.Suite
	store            *postgres.Store
	pool             *dockertest.Pool
	resource         *dockertest.Resource
	repository       *postgres.ApprovalRepository
	appealRepository *postgres.AppealRepository

	dummyProvider *domain.Provider
	dummyPolicy   *domain.Policy
	dummyResource *domain.Resource
	dummyAppeal   *domain.Appeal
}

func TestApprovalRepository(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, new(ApprovalRepositoryTestSuite))
}

func (s *ApprovalRepositoryTestSuite) SetupSuite() {
	var err error
	logger := log.NewCtxLogger("debug", []string{"test"})
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

	s.appealRepository = postgres.NewAppealRepository(s.store.DB())
	err = s.appealRepository.BulkUpsert(ctx, []*domain.Appeal{s.dummyAppeal})
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

func (s *ApprovalRepositoryTestSuite) TestGetApprovalsTotalCount() {
	s.Run("should return 0", func() {
		_, actualError := s.repository.GetApprovalsTotalCount(context.Background(), &domain.ListApprovalsFilter{})
		s.Nil(actualError)
	})
}

func (s *ApprovalRepositoryTestSuite) TestListApprovals() {
	pendingAppeal := &domain.Appeal{
		ResourceID:    s.dummyResource.ID,
		PolicyID:      s.dummyPolicy.ID,
		PolicyVersion: s.dummyPolicy.Version,
		AccountID:     "abc-user@example.com",
		AccountType:   domain.DefaultAppealAccountType,
		Role:          "role_test_a",
		Permissions:   []string{"permission_test"},
		CreatedBy:     "abc-user@example.com",
		Status:        domain.AppealStatusPending,
	}

	cancelledAppeal := &domain.Appeal{
		ResourceID:    s.dummyResource.ID,
		PolicyID:      s.dummyPolicy.ID,
		PolicyVersion: s.dummyPolicy.Version,
		AccountID:     "abc-user@example.com",
		AccountType:   domain.DefaultAppealAccountType,
		Role:          "role_test_b",
		Permissions:   []string{"permission_test"},
		CreatedBy:     "abc-user@example.com",
		Status:        domain.AppealStatusCanceled,
	}

	s.appealRepository.BulkUpsert(context.Background(), []*domain.Appeal{pendingAppeal, cancelledAppeal})

	dummyApprovals := []*domain.Approval{
		{
			Name:          "test-approval-name-1",
			Index:         0,
			AppealID:      s.dummyAppeal.ID,
			Status:        "test-status-1",
			PolicyID:      "test-policy-id",
			PolicyVersion: 1,
			Appeal:        s.dummyAppeal,
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
		{
			Name:          "test-approval-name-3",
			Index:         1,
			AppealID:      s.dummyAppeal.ID,
			Status:        "test-status-1",
			PolicyID:      "test-policy-id",
			PolicyVersion: 1,
			Appeal:        s.dummyAppeal,
		},
		{
			Name:          "test-approval-name-4",
			Index:         1,
			AppealID:      pendingAppeal.ID,
			Status:        "test-status-1",
			PolicyID:      "test-policy-id",
			PolicyVersion: 1,
			Appeal:        pendingAppeal,
		},
		{
			Name:          "test-approval-name-5",
			Index:         1,
			AppealID:      cancelledAppeal.ID,
			Status:        "test-status-1",
			PolicyID:      "test-policy-id",
			PolicyVersion: 1,
			Appeal:        cancelledAppeal,
		},
	}

	err := s.repository.BulkInsert(context.Background(), dummyApprovals)
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
		{
			ApprovalID: dummyApprovals[2].ID,
			AppealID:   s.dummyAppeal.ID,
			Email:      "approver1@email.com",
		},
		{
			ApprovalID: dummyApprovals[3].ID,
			AppealID:   pendingAppeal.ID,
			Email:      "approver3@email.com",
		},
		{
			ApprovalID: dummyApprovals[4].ID,
			AppealID:   cancelledAppeal.ID,
			Email:      "approver3@email.com",
		},
	}

	ctx := context.Background()

	for _, ap := range dummyApprover {
		err = s.repository.AddApprover(ctx, ap)
		s.Require().NoError(err)
	}

	s.Run("should return list of approvals on success", func() {
		approvals, err := s.repository.ListApprovals(context.Background(), &domain.ListApprovalsFilter{
			AccountID: s.dummyAppeal.AccountID,
			CreatedBy: dummyApprover[0].Email,
			Statuses:  []string{"test-status-1"},
			OrderBy:   []string{"status", "updated_at:desc", "created_at"},
			Size:      1,
		})

		s.NoError(err)
		s.Len(approvals, 1)
		s.Equal(dummyApprovals[0].ID, approvals[0].ID)
	})

	s.Run("should return approvals based on query search input", func() {
		approvals, err := s.repository.ListApprovals(context.Background(), &domain.ListApprovalsFilter{
			Q: "abc-user", // expected to match account_id: "abc-user@example.com"
		})

		s.NoError(err)
		s.Len(approvals, 1)
		s.Equal(pendingAppeal.ID, approvals[0].AppealID)
	})

	s.Run("should return list of approvals based account types filter", func() {
		approvals, err := s.repository.ListApprovals(context.Background(), &domain.ListApprovalsFilter{
			AccountTypes: []string{"x-account-type"}, // match 0 records
		})

		s.NoError(err)
		s.Len(approvals, 0)
	})

	s.Run("should return list of approvals based resource types filter", func() {
		approvals, err := s.repository.ListApprovals(context.Background(), &domain.ListApprovalsFilter{
			ResourceTypes: []string{"x-resource-type"}, // match 0 records
		})

		s.NoError(err)
		s.Len(approvals, 0)
	})

	s.Run("should return list of approvals where appeal status is canceled", func() {
		approvals, err := s.repository.ListApprovals(context.Background(), &domain.ListApprovalsFilter{
			AccountID:      cancelledAppeal.AccountID,
			CreatedBy:      dummyApprover[3].Email,
			AppealStatuses: []string{domain.AppealStatusCanceled},
			OrderBy:        []string{"status", "updated_at:desc", "created_at"},
		})

		s.NoError(err)
		s.Len(approvals, 1)
		s.Equal(dummyApprovals[4].ID, approvals[0].ID)
	})

	s.Run("should return list of approvals where appeal status is pending", func() {
		approvals, err := s.repository.ListApprovals(context.Background(), &domain.ListApprovalsFilter{
			AccountID:      pendingAppeal.AccountID,
			CreatedBy:      dummyApprover[3].Email,
			AppealStatuses: []string{domain.AppealStatusPending},
			OrderBy:        []string{"status", "updated_at:desc", "created_at"},
		})

		s.NoError(err)
		s.Len(approvals, 1)
		s.Equal(dummyApprovals[3].ID, approvals[0].ID)
	})

	s.Run("should return error if conditions invalid", func() {
		approvals, err := s.repository.ListApprovals(context.Background(), &domain.ListApprovalsFilter{
			AccountID: "",
			CreatedBy: "",
			Statuses:  []string{},
			OrderBy:   []string{},
		})

		s.Error(err)
		s.Nil(approvals)
	})

	s.Run("should return error if db execution returns an error on listing approvers", func() {
		approvals, err := s.repository.ListApprovals(context.Background(), &domain.ListApprovalsFilter{
			OrderBy: []string{"not-a-column"},
		})

		s.Error(err)
		s.Nil(approvals)
	})
}

func (s *ApprovalRepositoryTestSuite) TestListApprovals__Search() {
	s.Run("should pass grouped condition properly and return result accordingly", func() {
		dummyAppeals := []*domain.Appeal{
			{
				ResourceID:    s.dummyResource.ID,
				PolicyID:      s.dummyPolicy.ID,
				PolicyVersion: s.dummyPolicy.Version,
				AccountID:     "user1@example.com",
				AccountType:   domain.DefaultAppealAccountType,
				Role:          "role_test",
				Permissions:   []string{"permission_test"},
				CreatedBy:     "user1@example.com",
				Status:        domain.AppealStatusPending,
			},
			{
				ResourceID:    s.dummyResource.ID,
				PolicyID:      s.dummyPolicy.ID,
				PolicyVersion: s.dummyPolicy.Version,
				AccountID:     "user2@example.com",
				AccountType:   domain.DefaultAppealAccountType,
				Role:          "role_test",
				Permissions:   []string{"permission_test"},
				CreatedBy:     "user2@example.com",
				Status:        domain.AppealStatusPending,
			},
		}
		err := s.appealRepository.BulkUpsert(context.Background(), dummyAppeals)
		s.Require().NoError(err)

		dummyApprovals := []*domain.Approval{
			{
				Name:          "test-approval-name-1",
				Index:         0,
				AppealID:      dummyAppeals[0].ID,
				Status:        domain.ApprovalStatusPending,
				PolicyID:      "test-policy-id",
				PolicyVersion: 1,
				Appeal:        dummyAppeals[0],
			},
			{
				Name:          "test-approval-name-1",
				Index:         0,
				AppealID:      dummyAppeals[1].ID,
				Status:        domain.ApprovalStatusPending,
				PolicyID:      "test-policy-id",
				PolicyVersion: 1,
				Appeal:        dummyAppeals[1],
			},
		}
		err = s.repository.BulkInsert(context.Background(), dummyApprovals)
		s.Require().NoError(err)

		dummyApprover := []*domain.Approver{
			{
				ApprovalID: dummyApprovals[0].ID,
				AppealID:   dummyAppeals[0].ID,
				Email:      "approver@email.com",
			},
			{
				ApprovalID: dummyApprovals[1].ID,
				AppealID:   dummyAppeals[1].ID,
				Email:      "approver2@email.com",
			},
		}
		for _, ap := range dummyApprover {
			err = s.repository.AddApprover(context.Background(), ap)
			s.Require().NoError(err)
		}

		approvals, err := s.repository.ListApprovals(context.Background(), &domain.ListApprovalsFilter{
			CreatedBy: "approver@email.com",
			Q:         "role_test",
		})

		s.NoError(err)
		s.Len(approvals, 1)
		s.Equal(dummyApprovals[0].ID, approvals[0].ID)
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
		actualError := s.repository.BulkInsert(context.Background(), []*domain.Approval{})
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return nil error on success", func() {
		err := s.repository.BulkInsert(context.Background(), approvals)
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

	err := s.repository.BulkInsert(context.Background(), dummyApprovals)
	s.Require().NoError(err)

	s.Run("should return nil error on success", func() {
		dummyApprover := &domain.Approver{
			ApprovalID: dummyApprovals[0].ID,
			Email:      "user@example.com",
			AppealID:   s.dummyAppeal.ID,
		}

		err := s.repository.AddApprover(context.Background(), dummyApprover)
		s.NoError(err)
	})

	s.Run("should return error if approver payload is invalid", func() {
		invalidApprover := &domain.Approver{
			ID: "invalid-uuid",
		}

		err := s.repository.AddApprover(context.Background(), invalidApprover)

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

	ctx := context.Background()
	err := s.repository.BulkInsert(ctx, dummyApprovals)
	s.Require().NoError(err)

	dummyApprover := &domain.Approver{
		ApprovalID: dummyApprovals[0].ID,
		Email:      "user@example.com",
		AppealID:   s.dummyAppeal.ID,
	}

	err = s.repository.AddApprover(ctx, dummyApprover)
	s.NoError(err)

	s.Run("should return nil error on success", func() {
		err := s.repository.DeleteApprover(context.Background(), dummyApprovals[0].ID, dummyApprover.Email)
		s.NoError(err)
	})

	s.Run("should return error if db returns an error", func() {
		err := s.repository.DeleteApprover(context.Background(), "", "")
		s.Error(err)
	})
}
