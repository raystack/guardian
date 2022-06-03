package scheduler

import (
	"log"
	"runtime/debug"

	"github.com/robfig/cron/v3"
)

// Task defines a particular job needs to be done by the scheduler
type Task struct {
	CronTab string
	Func    func() error
}

// Scheduler is a cron wrapper that manages tasks
type Scheduler struct {
	cron *cron.Cron
}

// New returns new *Scheduler
func New(tasks []*Task) (*Scheduler, error) {
	c := cron.New()

	for _, t := range tasks {
		_, err := c.AddFunc(t.CronTab, handler(t.Func))
		if err != nil {
			return nil, err
		}
	}

	return &Scheduler{c}, nil
}

// Run starts scheduler to do the tasks
func (s *Scheduler) Run() {
	s.cron.Start()
}

func handler(fn func() error) func() {
	return func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("scheduled job error: %v\n%v", err, string(debug.Stack()))
			}
		}()

		if err := fn(); err != nil {
			log.Printf("scheduled job error: %v", err)
		}
	}
}
