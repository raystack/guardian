package domain

import "time"

type Namespace struct {
	ID        string
	Name      string
	State     string
	Metadata  map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NamespaceFilter struct {
}
