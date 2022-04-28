//go:generate mockery --name=repository --exported

package audit

import (
	"context"
	"time"
)

var TimeNow = time.Now

type repository interface {
	Init(context.Context) error
	Insert(context.Context, *Log) error
}

type AuditOption func(*Service)

func WithRepository(r repository) AuditOption {
	return func(s *Service) {
		s.repository = r
	}
}

func WithAppDetails(app AppDetails) AuditOption {
	return func(s *Service) {
		s.appDetails = app
	}
}

func WithTrackIDExtractor(fn func(ctx context.Context) string) AuditOption {
	return func(s *Service) {
		s.trackIDExtractor = fn
	}
}

type Service struct {
	appDetails       AppDetails
	repository       repository
	trackIDExtractor func(ctx context.Context) string
}

func New(opts ...AuditOption) *Service {
	svc := &Service{}
	for _, o := range opts {
		o(svc)
	}

	return svc
}

func (s *Service) Log(ctx context.Context, actor, action string, data interface{}) error {
	var traceID string
	if s.trackIDExtractor != nil {
		traceID = s.trackIDExtractor(ctx)
	}
	return s.repository.Insert(ctx, &Log{
		TraceID:   traceID,
		Timestamp: TimeNow(),
		Action:    action,
		Actor:     actor,
		Data:      data,
		App:       &s.appDetails,
	})
}
