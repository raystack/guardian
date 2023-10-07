package gcloudiam

const (
	ResourceTypeProject        = "project"
	ResourceTypeOrganization   = "organization"
	ResourceTypeServiceAccount = "service_account"
)

type Role struct {
	Name        string
	Title       string
	Description string
}
