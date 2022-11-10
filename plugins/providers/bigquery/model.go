package bigquery

import (
	"fmt"
	"strings"

	bq "cloud.google.com/go/bigquery"
	"github.com/odpf/guardian/domain"
)

const (
	// ResourceTypeDataset is the resource type name for BigQuery dataset
	ResourceTypeDataset = "dataset"
	// ResourceTypeTable is the resource type name for BigQuery table
	ResourceTypeTable = "table"
)

// Dataset is a reference to a BigQuery dataset
type Dataset struct {
	ProjectID string
	DatasetID string
}

func (d *Dataset) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeDataset {
		return ErrInvalidResourceType
	}

	d.ProjectID = strings.TrimSuffix(r.URN, fmt.Sprintf(":%s", r.Name))
	d.DatasetID = r.Name
	return nil
}

func (d *Dataset) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeDataset,
		Name: d.DatasetID,
		URN:  fmt.Sprintf("%s:%s", d.ProjectID, d.DatasetID),
	}
}

// Table is a reference to a BigQuery table
type Table struct {
	ProjectID string
	DatasetID string
	TableID   string
}

func (t *Table) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeTable {
		return ErrInvalidResourceType
	}

	datasetURN := strings.Split(strings.TrimSuffix(r.URN, fmt.Sprintf(".%s", r.Name)), ":")
	if len(datasetURN) != 2 {
		return ErrInvalidTableURN
	}
	t.ProjectID = datasetURN[0]
	t.DatasetID = datasetURN[1]
	t.TableID = r.Name
	return nil
}

func (t *Table) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeTable,
		Name: t.TableID,
		URN:  fmt.Sprintf("%s:%s.%s", t.ProjectID, t.DatasetID, t.TableID),
	}
}

type datasetAccessEntry bq.AccessEntry

func (ae datasetAccessEntry) getEntityType() string {
	switch bq.AccessEntry(ae).EntityType {
	case bq.UserEmailEntity:
		if strings.Contains(bq.AccessEntry(ae).Entity, "gserviceaccount.com") {
			return AccountTypeServiceAccount
		}
		return AccountTypeUser
	default:
		return ""
	}
}

func resolvePermissions(_permissions []interface{}) []interface{} {
	var permissions []interface{}
	for _, p := range _permissions {
		role := p.(string)
		switch role {
		case "roles/bigquery.dataViewer":
			permissions = append(permissions, bq.ReaderRole)
		case "roles/bigquery.dataOwner":
			permissions = append(permissions, bq.OwnerRole)
		case "roles/bigquery.dataEditor":
			permissions = append(permissions, bq.WriterRole)
		default:
			permissions = append(permissions, bq.AccessRole(role))
		}
	}
	return permissions
}
