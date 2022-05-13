package audit

import "time"

type AppDetails struct {
	Name    string
	Version string
}

type Log struct {
	Timestamp time.Time
	Action    string
	Actor     string
	App       *AppDetails
	Data      interface{}
	Metadata  map[string]interface{}
}
