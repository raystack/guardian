package grafana

import (
	"strconv"

	"github.com/goto/guardian/domain"
)

const (
	ResourceTypeFolder    = "folder"
	ResourceTypeDashboard = "dashboard"
)

type Folder struct {
	ID    int    `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
}

type Dashboard struct {
	ID          int    `json:"id"`
	UID         string `json:"uid"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	FolderID    int    `json:"folderId"`
	FolderUID   string `json:"folderUid"`
	FolderTitle string `json:"folderTitle"`
}

func (d *Dashboard) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeDashboard {
		return ErrInvalidResourceType
	}

	id, err := strconv.Atoi(r.URN)
	if err != nil {
		return err
	}

	d.ID = id
	d.Title = r.Name
	return nil
}

func (d *Dashboard) ToDomain() *domain.Resource {
	details := map[string]interface{}{}
	id := strconv.Itoa(d.ID)

	if d.FolderID != 0 {
		details["folder_id"] = d.FolderID
	}
	if d.FolderUID != "" {
		details["folder_uid"] = d.FolderUID
	}
	if d.FolderTitle != "" {
		details["folder_title"] = d.FolderTitle
	}
	if d.UID != "" {
		details["uid"] = d.UID
	}
	return &domain.Resource{
		Type:    ResourceTypeDashboard,
		Name:    d.Title,
		URN:     id,
		Details: details,
	}
}
