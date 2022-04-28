//go:generate mockery --name=repository --exported

package audit

import (
	"context"
	"time"
)

var TimeNow = time.Now

type actorContextKey struct{}

func WithActor(ctx context.Context, actor string) context.Context {
	return context.WithValue(ctx, actorContextKey{}, actor)
}

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

func (s *Service) Log(ctx context.Context, action string, data interface{}) error {
	l := &Log{
		Timestamp: TimeNow(),
		Action:    action,
		Data:      data,
		App:       &s.appDetails,
	}

	if s.trackIDExtractor != nil {
		l.TraceID = s.trackIDExtractor(ctx)
	}

	if actor, ok := ctx.Value(actorContextKey{}).(string); ok {
		l.Actor = actor
	}

	return s.repository.Insert(ctx, l)
}
