package shield

import (
	"fmt"
	"github.com/odpf/guardian/domain"
)

const (
	ResourceTypeTeam         = "team"
	ResourceTypeProject      = "project"
	ResourceTypeOrganization = "organization"
)

type Metadata struct {
	Email   string `json:"email" mapstructure:"email"`
	Privacy string `json:"privacy" mapstructure:"privacy"`
	Slack   string `json:"slack" mapstructure:"slack"`
}

type User struct {
	ID    string `json:"id" mapstructure:"id"`
	Name  string `json:"name" mapstructure:"name"`
	Email string `json:"email" mapstructure:"email"`
}

type Team struct {
	ID       string   `json:"id" mapstructure:"id"`
	Name     string   `json:"name" mapstructure:"name"`
	Slug     string   `json:"slug" mapstructure:"slug"`
	OrgId    string   `json:"orgId" mapstructure:"orgId"`
	Metadata Metadata `json:"metadata" mapstructure:"metadata"`
	Admins   []string `json:"admin" mapstructure:"admin"`
}

type Project struct {
	ID     string   `json:"id" mapstructure:"id"`
	Name   string   `json:"name" mapstructure:"name"`
	Slug   string   `json:"slug" mapstructure:"slug"`
	OrgId  string   `json:"orgId" mapstructure:"orgId"`
	Admins []string `json:"admin" mapstructure:"admin"`
}

type Organization struct {
	ID     string   `json:"id" mapstructure:"id"`
	Name   string   `json:"name" mapstructure:"name"`
	Slug   string   `json:"slug" mapstructure:"slug"`
	Admins []string `json:"admin" mapstructure:"admin"`
}

func (t *Team) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeTeam {
		return ErrInvalidResourceType
	}

	resourseDetails := r.Details
	t.ID = resourseDetails["id"].(string)
	t.OrgId = resourseDetails["orgId"].(string)
	t.Metadata = resourseDetails["metadata"].(Metadata)
	t.Admins = resourseDetails["admins"].([]string)
	t.Name = r.Name

	return nil
}

func (t *Team) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeTeam,
		Name: t.Name,
		URN:  fmt.Sprintf("team:%v", t.ID),
		Details: map[string]interface{}{
			"id":       t.ID,
			"metadata": t.Metadata,
			"orgId":    t.OrgId,
			"admins":   t.Admins,
		},
	}
}

func (p *Project) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeProject {
		return ErrInvalidResourceType
	}

	resourseDetails := r.Details
	p.ID = resourseDetails["id"].(string)
	p.OrgId = resourseDetails["orgId"].(string)
	p.Admins = resourseDetails["admins"].([]string)
	p.Name = r.Name

	return nil
}

func (p *Project) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeProject,
		Name: p.Name,
		URN:  fmt.Sprintf("project:%v", p.ID),
		Details: map[string]interface{}{
			"id":     p.ID,
			"orgId":  p.OrgId,
			"admins": p.Admins,
		},
	}
}

func (o *Organization) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeOrganization {
		return ErrInvalidResourceType
	}

	resourseDetails := r.Details
	o.ID = resourseDetails["id"].(string)
	o.Admins = resourseDetails["admins"].([]string)
	o.Name = r.Name

	return nil
}

func (o *Organization) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeOrganization,
		Name: o.Name,
		URN:  fmt.Sprintf("organization:%v", o.ID),
		Details: map[string]interface{}{
			"id":     o.ID,
			"admins": o.Admins,
		},
	}
}
