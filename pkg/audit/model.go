package audit

import "time"

type AppDetails struct {
	Name    string
	Version string
}

type Log struct {
	TraceID   string
	Timestamp time.Time
	Action    string
	Actor     string
	App       *AppDetails
	Data      interface{}
}
