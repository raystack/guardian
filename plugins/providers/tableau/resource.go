package tableau

import (
	"time"

	"github.com/odpf/guardian/domain"
)

const (
	ResourceTypeWorkbook   = "workbook"
	ResourceTypeFlow       = "flow"
	ResourceTypeDataSource = "datasource"
	ResourceTypeView       = "view"
	ResourceTypeMetric     = "metric"
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

type Flow struct {
	Project    project     `json:"project"`
	Owner      owner       `json:"owner"`
	Tags       interface{} `json:"tags"`
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	WebpageURL string      `json:"webpageUrl"`
	FileType   string      `json:"fileType"`
}

type DataSource struct {
	Project             project     `json:"project"`
	Owner               owner       `json:"owner"`
	Tags                interface{} `json:"tags"`
	ID                  string      `json:"id"`
	Name                string      `json:"name"`
	EncryptExtracts     string      `json:"encryptExtracts"`
	ContentURL          string      `json:"contentUrl"`
	HasExtracts         bool        `json:"hasExtracts"`
	IsCertified         bool        `json:"isCertified"`
	Type                string      `json:"type"`
	UseRemoteQueryAgent bool        `json:"useRemoteQueryAgent"`
	WebpageURL          string      `json:"webpageUrl"`
}

type View struct {
	Project     project     `json:"project"`
	Owner       owner       `json:"owner"`
	Workbook    workbook    `json:"workbook"`
	Tags        interface{} `json:"tags"`
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	ContentURL  string      `json:"contentUrl"`
	ViewUrlName string      `json:"viewUrlName"`
}

type Metric struct {
	Project        project        `json:"project"`
	Owner          owner          `json:"owner"`
	Tags           interface{}    `json:"tags"`
	UnderlyingView UnderlyingView `json:"underlyingView"`
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	WebpageURL     string         `json:"webpageUrl"`
	Suspended      bool           `json:"suspended"`
}

type project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type owner struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UnderlyingView struct {
	ID string `json:"id"`
}

type workbook struct {
	ID string `json:"id"`
}

type dataAccelerationConfig struct {
	AccelerationEnabled bool `json:"accelerationEnabled"`
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

func (f *Flow) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeFlow {
		return ErrInvalidResourceType
	}

	f.ID = r.URN
	f.Name = r.Name
	return nil
}

func (f *Flow) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeFlow,
		Name: f.Name,
		URN:  f.ID,
		Details: map[string]interface{}{
			"project_name": f.Project.Name,
			"project_id":   f.Project.ID,
			"owner_id":     f.Owner.ID,
			"webpage_url":  f.WebpageURL,
			"tags":         f.Tags,
			"fileType":     f.FileType,
		},
	}
}

func (d *DataSource) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeDataSource {
		return ErrInvalidResourceType
	}

	d.ID = r.URN
	d.Name = r.Name
	return nil
}

func (d *DataSource) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeDataSource,
		Name: d.Name,
		URN:  d.ID,
		Details: map[string]interface{}{
			"project_name":        d.Project.Name,
			"project_id":          d.Project.ID,
			"owner_id":            d.Owner.ID,
			"content_url":         d.ContentURL,
			"webpage_url":         d.WebpageURL,
			"tags":                d.Tags,
			"encryptExtracts":     d.EncryptExtracts,
			"hasExtracts":         d.HasExtracts,
			"isCertified":         d.IsCertified,
			"type":                d.Type,
			"useRemoteQueryAgent": d.UseRemoteQueryAgent,
		},
	}
}

func (v *View) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeView {
		return ErrInvalidResourceType
	}

	v.ID = r.URN
	v.Name = r.Name
	return nil
}

func (v *View) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeView,
		Name: v.Name,
		URN:  v.ID,
		Details: map[string]interface{}{
			"project_name": v.Project.Name,
			"project_id":   v.Project.ID,
			"owner_name":   v.Owner.Name,
			"workbook_id":  v.Workbook.ID,
			"owner_id":     v.Owner.ID,
			"content_url":  v.ContentURL,
			"tags":         v.Tags,
			"viewUrlName":  v.ViewUrlName,
		},
	}
}

func (m *Metric) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeMetric {
		return ErrInvalidResourceType
	}

	m.ID = r.URN
	m.Name = r.Name
	return nil
}

func (m *Metric) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeMetric,
		Name: m.Name,
		URN:  m.ID,
		Details: map[string]interface{}{
			"project_name":   m.Project.Name,
			"project_id":     m.Project.ID,
			"owner_id":       m.Owner.ID,
			"webpage_url":    m.WebpageURL,
			"tags":           m.Tags,
			"description":    m.Description,
			"underlyingView": m.UnderlyingView,
			"suspended":      m.Suspended,
		},
	}
}
