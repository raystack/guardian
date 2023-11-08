package log

import (
	"context"
	"io"
	"io/ioutil"
)

type Noop struct{}
type Option func(interface{})

func (n *Noop) Debug(ctx context.Context, msg string, args ...interface{}) {}
func (n *Noop) Info(ctx context.Context, msg string, args ...interface{})  {}
func (n *Noop) Warn(ctx context.Context, msg string, args ...interface{})  {}
func (n *Noop) Error(ctx context.Context, msg string, args ...interface{}) {}
func (n *Noop) Fatal(ctx context.Context, msg string, args ...interface{}) {}

func (n *Noop) Level() string {
	return "unsupported"
}
func (n *Noop) Writer() io.Writer {
	return ioutil.Discard
}

// NewNoop returns a no operation logger, useful in tests
// to avoid printing logs to stdout.
func NewNoop(opts ...Option) *Noop {
	return &Noop{}
}
