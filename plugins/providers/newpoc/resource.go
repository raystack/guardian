package newpoc

import "fmt"

const (
	ResourceTypeProject      = "project"
	ResourceTypeOrganization = "organization"
)

type resource struct {
	Type string
	ID   string
}

func (r resource) GetType() string {
	return r.Type
}

func (r resource) GetURN() string {
	return fmt.Sprintf("%s/%s", r.Type, r.ID)
}

func (r resource) GetDisplayName() string {
	return fmt.Sprintf("%s - GCP IAM", r.GetURN())
}

func (r resource) GetMetadata() map[string]interface{} {
	return nil
}
