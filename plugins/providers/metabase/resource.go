package metabase

import (
	"fmt"
	"strconv"

	"github.com/odpf/guardian/domain"
)

const (
	ResourceTypeDatabase   = "database"
	ResourceTypeCollection = "collection"
)

type Database struct {
	ID                       int    `json:"id"`
	Name                     string `json:"name"`
	CacheFieldValuesSchedule string `json:"cache_field_values_schedule"`
	Timezone                 string `json:"timezone"`
	AutoRunQueries           bool   `json:"auto_run_queries"`
	MetadataSyncSchedule     string `json:"metadata_sync_schedule"`
	Engine                   string `json:"engine"`
	NativePermissions        string `json:"native_permissions"`
}

func (d *Database) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeDatabase {
		return ErrInvalidResourceType
	}

	id, err := strconv.Atoi(r.URN)
	if err != nil {
		return err
	}

	d.ID = id
	d.Name = r.Name
	return nil
}

func (d *Database) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeDatabase,
		Name: d.Name,
		URN:  fmt.Sprintf("%v", d.ID),
		Details: map[string]interface{}{
			"cache_field_values_schedule": d.CacheFieldValuesSchedule,
			"timezone":                    d.Timezone,
			"auto_run_queries":            d.AutoRunQueries,
			"metadata_sync_schedule":      d.MetadataSyncSchedule,
			"engine":                      d.Engine,
			"native_permissions":          d.NativePermissions,
		},
	}
}

type Collection struct {
	ID        interface{} `json:"id"`
	Name      string      `json:"name"`
	Slug      string      `json:"slug"`
	Location  string      `json:"location,omitempty"`
	Namespace string      `json:"namespace,omitempty"`
}

func (c *Collection) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeCollection {
		return ErrInvalidResourceType
	}

	id, _ := strconv.Atoi(r.URN)
	if id == 0 {
		c.ID = r.URN
	} else {
		c.ID = id
	}
	c.Name = r.Name
	return nil
}

func (c *Collection) ToDomain() *domain.Resource {
	details := map[string]interface{}{}
	if c.Location != "" {
		details["location"] = c.Location
	}
	if c.Namespace != "" {
		details["namespace"] = c.Namespace
	}
	return &domain.Resource{
		Type:    ResourceTypeCollection,
		Name:    c.Name,
		URN:     fmt.Sprintf("%v", c.ID),
		Details: details,
	}
}
