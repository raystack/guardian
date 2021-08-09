package tableau

import (
	"time"

	"github.com/odpf/guardian/domain"
)

const (
	ResourceTypeWorkbook = "workbook"
	ResourceTypeFlow     = "flow"
)

type Workbook struct {
	Project                project                `json:"project"`
	Owner                  owner                  `json:"owner"`
	Tags                   interface{}            `json:"tags"`
	DataAccelerationConfig dataAccelerationConfig `json:"dataAccelerationConfig"`
	ID                     string                 `json:"id"`
	Name                   string                 `json:"name"`
	ContentURL             string                 `json:"contentUrl"`
	WebpageURL             string                 `json:"webpageUrl"`
	ShowTabs               string                 `json:"showTabs"`
	Size                   string                 `json:"size"`
	CreatedAt              time.Time              `json:"createdAt"`
	UpdatedAt              time.Time              `json:"updatedAt"`
	EncryptExtracts        string                 `json:"encryptExtracts"`
	DefaultViewID          string                 `json:"defaultViewId"`
}

func (w *Workbook) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeWorkbook {
		return ErrInvalidResourceType
	}

	w.ID = r.URN
	w.Name = r.Name
	return nil
}

func (w *Workbook) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeWorkbook,
		Name: w.Name,
		URN:  w.ID,
		Details: map[string]interface{}{
			"project_name":    w.Project.Name,
			"project_id":      w.Project.ID,
			"owner_name":      w.Owner.Name,
			"owner_id":        w.Owner.ID,
			"content_url":     w.ContentURL,
			"webpage_url":     w.WebpageURL,
			"size":            w.Size,
			"default_view_id": w.DefaultViewID,
			"tags":            w.Tags,
			"show_tabs":       w.ShowTabs,
		},
	}
}

type project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type owner struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type dataAccelerationConfig struct {
	AccelerationEnabled bool `json:"accelerationEnabled"`
}
