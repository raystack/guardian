package grafana

import (
	"github.com/odpf/guardian/domain"
)

const (
	ResourceTypeFolder    = "folder"
	ResourceTypeDashboard = "dashboard"
)

type Folder struct {
	ID    uint   `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
}

func (f *Folder) toDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeFolder,
		Name: f.Title,
		ID:   f.ID,
		URN:  f.UID,
	}
}

type Dashboard struct {
	ID          uint   `json:"id"`
	UID         string `json:"uid"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	FolderID    uint   `json:"folderId"`
	FolderUId   string `json:"folderUid"`
	FolderTitle string `json:"folderTitle"`
}

func (d *Dashboard) toDomain() *domain.Resource {
	details := map[string]interface{}{}
	if d.FolderID != 0 {
		details["folderId"] = d.FolderID
	}
	if d.FolderUId != "" {
		details["folderUid"] = d.FolderUId
	}
	if d.FolderTitle != "" {
		details["folderTitle"] = d.FolderTitle
	}
	return &domain.Resource{
		Type:    ResourceTypeDashboard,
		Name:    d.Title,
		URN:     d.UID,
		ID:      d.ID,
		Details: details,
	}
}
