package audit

import (
	"context"
	"time"
)

type repository interface {
	Insert(context.Context, *Log) error
}

type AuditOption func(*service)

func WithRepository(r repository) AuditOption {
	return func(s *service) {
		s.repository = r
	}
}

func WithAppDetails(app AppDetails) AuditOption {
	return func(s *service) {
		s.appDetails = app
	}
}

func WithTrackIDExtractor(fn func(ctx context.Context) string) AuditOption {
	return func(s *service) {
		s.trackIDExtractor = fn
	}
}

type service struct {
	appDetails       AppDetails
	repository       repository
	trackIDExtractor func(ctx context.Context) string
}

func New(opts ...AuditOption) *service {
	svc := &service{}
	for _, o := range opts {
		o(svc)
	}

	if svc.repository == nil {
		svc.repository = NewLogRepository()
	}

	return svc
}

func (s *service) Log(ctx context.Context, actor, action string, data interface{}) error {
	var traceID string
	if s.trackIDExtractor != nil {
		traceID = s.trackIDExtractor(ctx)
	}
	return s.repository.Insert(ctx, &Log{
		TraceID:   traceID,
		Timestamp: time.Now(),
		Action:    action,
		Actor:     actor,
		Data:      data,
		App:       &s.appDetails,
	})
}
