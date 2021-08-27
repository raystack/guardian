package gcloudiam

const (
	ResourceTypeProject      = "project"
	ResourceTypeOrganization = "organization"
)

type Role struct {
	Name        string
	Title       string
	Description string
}
