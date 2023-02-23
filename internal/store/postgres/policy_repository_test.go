package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/odpf/guardian/core/policy"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/store/postgres"
	"github.com/odpf/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type PolicyRepositoryTestSuite struct {
	suite.Suite
	store      *postgres.Store
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.PolicyRepository
}

func (s *PolicyRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewLogrus(log.LogrusWithLevel("debug"))
	s.store, s.pool, s.resource, err = newTestStore(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.repository = postgres.NewPolicyRepository(s.store.DB())
}

func (s *PolicyRepositoryTestSuite) TearDownSuite() {
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

func (s *PolicyRepositoryTestSuite) TestCreate() {
	s.Run("should return error if payload is invalid", func() {
		p := &domain.Policy{
			IAM: &domain.IAMConfig{
				Config: make(chan int),
			},
			AppealConfig: &domain.PolicyAppealConfig{
				AllowPermanentAccess:         false,
				AllowActiveAccessExtensionIn: "24h",
				Questions: []domain.Question{
					{
						Key:         "team",
						Question:    "What team are you in?",
						Required:    true,
						Description: "Please provide the name of the team you are in",
					},
					{
						Key:      "purpose",
						Question: "What is the purpose of this access?",
						Required: false,
						Description: "Explain why you need this access. " +
							"This will be used to evaluate your appeal. " +
							"For example, you may need access to a specific project or resource.",
					},
				},
			},
		}
		actualError := s.repository.Create(context.Background(), p)

		s.EqualError(actualError, "serializing policy: json: unsupported type: chan int")
	})

	s.Run("should return nil error on success", func() {
		p := &domain.Policy{
			ID: "test_policy",
		}
		err := s.repository.Create(context.Background(), p)
		s.Nil(err)
		s.NotEmpty(p.ID)
	})
}

func (s *PolicyRepositoryTestSuite) TestFind() {
	err1 := setup(s.store)
	s.Nil(err1)

	s.Run("should return list of policies on success", func() {
		ctx := context.Background()
		expectedPolicies := []*domain.Policy{
			{
				Version:     1,
				Description: "test_policy",
				AppealConfig: &domain.PolicyAppealConfig{
					AllowPermanentAccess:         false,
					AllowActiveAccessExtensionIn: "24h",
					Questions: []domain.Question{
						{
							Key:         "team",
							Question:    "What team are you in?",
							Required:    true,
							Description: "Please provide the name of the team you are in",
						},
						{
							Key:      "purpose",
							Question: "What is the purpose of this access?",
							Required: false,
							Description: "Explain why you need this access. " +
								"This will be used to evaluate your appeal. " +
								"For example, you may need access to a specific project or resource.",
						},
					},
				},
			},
		}

		for _, pol := range expectedPolicies {
			err := s.repository.Create(ctx, pol)
			s.Nil(err)
		}

		actualPolicies, actualError := s.repository.Find(ctx)

		if diff := cmp.Diff(expectedPolicies, actualPolicies, cmpopts.EquateApproxTime(time.Microsecond)); diff != "" {
			s.T().Errorf("result not match, diff: %v", diff)
		}
		s.Nil(actualError)
	})
}

func (s *PolicyRepositoryTestSuite) TestGetOne() {
	err1 := setup(s.store)
	s.Nil(err1)

	s.Run("should return error if record not found", func() {
		expectedError := policy.ErrPolicyNotFound

		sampleUUID := uuid.New().String()
		actualResult, actualError := s.repository.GetOne(context.Background(), sampleUUID, 0)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should pass args based on the version param", func() {
		dummyPolicies := []*domain.Policy{
			{
				ID:      "test-id",
				Version: 0,
			},
			{
				ID:      "test-id",
				Version: 1,
			},
			{
				ID:      "test-id",
				Version: 2,
			},
		}
		for _, p := range dummyPolicies {
			err := s.repository.Create(context.Background(), p)
			s.Require().NoError(err)
		}

		testCases := []struct {
			name            string
			versionParam    uint
			expectedVersion uint
		}{
			{
				name:            "should return latest version if version param is empty",
				expectedVersion: 2,
			},
			{
				name:            "should return expected version",
				versionParam:    1,
				expectedVersion: 1,
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				actualPolicy, actualError := s.repository.GetOne(context.Background(), "test-id", tc.versionParam)

				s.NoError(actualError)
				s.Equal(tc.expectedVersion, actualPolicy.Version)
			})
		}
	})
}

func TestPolicyRepository(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, new(PolicyRepositoryTestSuite))
}
