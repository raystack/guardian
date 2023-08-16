package bigquery

import (
	"fmt"
	"strings"

	bq "cloud.google.com/go/bigquery"
	"github.com/goto/guardian/core/resource"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/utils"
	"github.com/mitchellh/mapstructure"
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
	Labels    map[string]string
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
	r := &domain.Resource{
		Type: ResourceTypeDataset,
		Name: d.DatasetID,
		URN:  fmt.Sprintf("%s:%s", d.ProjectID, d.DatasetID),
	}

	if d.Labels != nil {
		if r.Details == nil {
			r.Details = make(map[string]interface{})
		}
		r.Details[resource.ReservedDetailsKeyMetadata] = map[string]interface{}{
			"labels": d.Labels,
		}
	}

	return r
}

// Table is a reference to a BigQuery table
type Table struct {
	ProjectID string
	DatasetID string
	TableID   string
	Labels    map[string]string
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
	r := &domain.Resource{
		Type: ResourceTypeTable,
		Name: t.TableID,
		URN:  fmt.Sprintf("%s:%s.%s", t.ProjectID, t.DatasetID, t.TableID),
	}

	if t.Labels != nil {
		if r.Details == nil {
			r.Details = make(map[string]interface{})
		}
		r.Details[resource.ReservedDetailsKeyMetadata] = map[string]interface{}{
			"labels": t.Labels,
		}
	}

	return r
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

// BigQueryResourceName is a string representation of bigquery resource's Relative Resource Name.
// Example: "projects/project-id/datasets/dataset_name/tables/table_name"
type BigQueryResourceName string

func (r BigQueryResourceName) ProjectID() string {
	items := strings.Split(string(r), "/")
	if len(items) >= 2 {
		return items[1]
	}
	return ""
}

func (r BigQueryResourceName) DatasetID() string {
	items := strings.Split(string(r), "/")
	if len(items) >= 4 {
		return items[3]
	}
	return ""
}

func (r BigQueryResourceName) TableID() string {
	items := strings.Split(string(r), "/")
	if len(items) >= 6 {
		return items[5]
	}
	return ""
}

// BigQueryResourceID returns bigquery resource identifier in format of:
// For dataset type: "project-id:dataset_name"
// For table type: "project-id:dataset_name.table_name"
func (r BigQueryResourceName) BigQueryResourceID() string {
	urn := fmt.Sprintf("%s:%s", r.ProjectID(), r.DatasetID())
	if tableID := r.TableID(); tableID != "" {
		urn = fmt.Sprintf("%s.%s", urn, tableID)
	}
	return urn
}

type cloudLoggingOptions struct {
	LogBucket string `json:"log_bucket" yaml:"log_bucket" mapstructure:"log_bucket"`
}

type activityConfig struct {
	*domain.ActivityConfig
}

func (c activityConfig) Validate() error {
	validSources := []string{"default", "cloud_logging"}
	if !utils.ContainsString(validSources, c.ActivityConfig.Source) {
		return fmt.Errorf("invalid source: %q, allowed values %v", c.ActivityConfig.Source, validSources)
	}

	return nil
}

func (c activityConfig) GetCloudLoggingOptions() (*cloudLoggingOptions, error) {
	if c.ActivityConfig == nil || (c.ActivityConfig.Source == "cloud_logging" && c.ActivityConfig.Options == nil) {
		return &cloudLoggingOptions{}, nil
	}
	if c.ActivityConfig.Source != "cloud_logging" {
		return nil, fmt.Errorf("invalid source: %q", c.ActivityConfig.Source)
	}

	result := &cloudLoggingOptions{}
	if err := mapstructure.Decode(c.ActivityConfig.Options, result); err != nil {
		return nil, fmt.Errorf("decoding options: %w", err)
	}
	return result, nil
}
