package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/odpf/guardian/core/policy"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PolicyRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	store      *Store
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *PolicyRepository
}

func (s *PolicyRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewLogrus(log.LogrusWithLevel("debug"))
	s.store, s.pool, s.resource, err = newTestStore(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = NewPolicyRepository(s.store.DB())
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
		}
		actualError := s.repository.Create(p)

		s.EqualError(actualError, "serializing p: json: unsupported type: chan int")
	})

	s.Run("should return nil error on success", func() {
		p := &domain.Policy{}
		err := s.repository.Create(p)
		s.Nil(err)
		s.NotEmpty(p.ID)
	})
}

func (s *PolicyRepositoryTestSuite) TestFind() {

	err1 := setup(s.store)
	s.Nil(err1)

	s.Run("should return list of policies on success", func() {
		expectedPolicies := []*domain.Policy{
			{
				Version:      1,
				Description:  "test_policy",
				AppealConfig: nil,
			},
		}

		for _, pol := range expectedPolicies {
			err := s.repository.Create(pol)
			s.Nil(err)
		}

		actualPolicies, actualError := s.repository.Find()

		s.Equal(expectedPolicies, actualPolicies)
		s.Nil(actualError)
	})
}

func (s *PolicyRepositoryTestSuite) TestGetOne() {
	err1 := setup(s.store)
	s.Nil(err1)

	s.Run("should return error if record not found", func() {
		expectedError := policy.ErrPolicyNotFound

		sampleUUID := uuid.New().String()
		actualResult, actualError := s.repository.GetOne(sampleUUID, 0)

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	//TODO: fix test case
	/*
		s.Run("should pass args based on the version param", func() {
			testCases := []struct {
				name            string
				expectedID      string
				expectedVersion uint
				expectedQuery   string
				expectedArgs    []driver.Value
			}{
				{
					name:            "should not apply version condition if version param given is 0",
					expectedID:      "test-id",
					expectedVersion: 0,
					expectedQuery:   regexp.QuoteMeta(`SELECT * FROM "policies" WHERE id = $1 AND "policies"."deleted_at" IS NULL ORDER BY version desc,"policies"."id" LIMIT 1`),
					expectedArgs:    []driver.Value{"test-id"},
				},
				{
					name:            "should apply version condition if version param is exists",
					expectedID:      "test-id",
					expectedVersion: 1,
					expectedQuery:   regexp.QuoteMeta(`SELECT * FROM "policies" WHERE (id = $1 AND version = $2) AND "policies"."deleted_at" IS NULL ORDER BY version desc,"policies"."id" LIMIT 1`),
					expectedArgs:    []driver.Value{"test-id", 1},
				},
			}

			for _, tc := range testCases {
				s.Run(tc.name, func() {
					now := time.Now()
					expectedRowValues := []driver.Value{
						tc.expectedID,
						tc.expectedVersion,
						"",
						"null",
						"null",
						"null",
						"null",
						"null",
						now,
						now,
					}
					s.dbmock.ExpectQuery(tc.expectedQuery).
						WithArgs(tc.expectedArgs...).
						WillReturnRows(sqlmock.NewRows(s.rows).AddRow(expectedRowValues...))

					_, actualError := s.repository.GetOne(tc.expectedID, tc.expectedVersion)

					s.Nil(actualError)
					s.dbmock.ExpectationsWereMet()
				})
			}
		})

	*/
}

func TestPolicyRepository(t *testing.T) {
	suite.Run(t, new(PolicyRepositoryTestSuite))
}
