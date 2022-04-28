package audit_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/guardian/pkg/audit"
	"github.com/odpf/guardian/pkg/audit/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AuditTestSuite struct {
	suite.Suite

	now time.Time

	mockRepository *mocks.Repository
	service        *audit.Service
}

func (s *AuditTestSuite) setupTest() {
	s.mockRepository = new(mocks.Repository)
	s.service = audit.New(
		audit.WithAppDetails(audit.AppDetails{Name: "guardian_test", Version: "1"}),
		audit.WithRepository(s.mockRepository),
		audit.WithTrackIDExtractor(func(_ context.Context) string {
			return "test-trace-id"
		}),
	)

	s.now = time.Now()
	audit.TimeNow = func() time.Time {
		return s.now
	}
}

func TestAudit(t *testing.T) {
	suite.Run(t, new(AuditTestSuite))
}

func (s *AuditTestSuite) TestLog() {
	s.Run("should insert to repository", func() {
		s.setupTest()

		s.mockRepository.On("Insert", mock.Anything, &audit.Log{
			TraceID:   "test-trace-id",
			Timestamp: s.now,
			Action:    "action",
			Actor:     "user@example.com",
			Data:      map[string]interface{}{"foo": "bar"},
			App: &audit.AppDetails{
				Name:    "guardian_test",
				Version: "1",
			},
		}).Return(nil)

		ctx := context.Background()
		ctx = audit.WithActor(ctx, "user@example.com")
		err := s.service.Log(ctx, "action", map[string]interface{}{"foo": "bar"})
		s.NoError(err)
	})

	s.Run("should pass empty trace id if extractor not found", func() {
		s.service = audit.New(
			audit.WithAppDetails(audit.AppDetails{Name: "guardian_test", Version: "1"}),
			audit.WithRepository(s.mockRepository),
		)

		s.mockRepository.On("Insert", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			l := args.Get(1).(*audit.Log)
			s.Empty(l.TraceID)
		}).Return(nil)

		err := s.service.Log(context.Background(), "", nil)
		s.NoError(err)
	})

	s.Run("should return error if repository.Insert fails", func() {
		s.setupTest()

		expectedError := errors.New("test error")
		s.mockRepository.On("Insert", mock.Anything, mock.Anything).Return(expectedError)

		err := s.service.Log(context.Background(), "", nil)
		s.ErrorIs(err, expectedError)
	})
}
