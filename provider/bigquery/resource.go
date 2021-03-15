package bigquery

import (
	"fmt"

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

func (d *Dataset) toDomain() *domain.Resource {
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

func (t *Table) toDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeTable,
		Name: t.TableID,
		URN:  fmt.Sprintf("%s:%s.%s", t.ProjectID, t.DatasetID, t.TableID),
	}
}
