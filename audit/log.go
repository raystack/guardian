package audit

import (
	"context"
	"fmt"
	"io"
	"log"
)

type logRepository struct {
	w io.Writer
}

func NewLogRepository() *logRepository {
	return &logRepository{
		w: log.Default().Writer(),
	}
}

func (r *logRepository) Insert(ctx context.Context, l *Log) error {
	str := fmt.Sprintf("%+v", l)
	_, err := r.w.Write([]byte(str))
	return err
}
