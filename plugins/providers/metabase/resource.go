package metabase

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/raystack/guardian/domain"
)

const (
	ResourceTypeDatabase   = "database"
	ResourceTypeTable      = "table"
	ResourceTypeCollection = "collection"
	ResourceTypeGroup      = "group"

	GuardianGroupPrefix = "_guardian_"
)

type Database struct {
	ID                       int     `json:"id" mapstructure:"id"`
	Name                     string  `json:"name" mapstructure:"name"`
	CacheFieldValuesSchedule string  `json:"cache_field_values_schedule" mapstructure:"cache_field_values_schedule"`
	Timezone                 string  `json:"timezone" mapstructure:"timezone"`
	AutoRunQueries           bool    `json:"auto_run_queries" mapstructure:"auto_run_queries"`
	MetadataSyncSchedule     string  `json:"metadata_sync_schedule" mapstructure:"metadata_sync_schedule"`
	Engine                   string  `json:"engine" mapstructure:"engine"`
	NativePermissions        string  `json:"native_permissions" mapstructure:"native_permissions"`
	Tables                   []Table `json:"tables" mapstructure:"tables"`
}

type Table struct {
	ID       int              `json:"id" mapstructure:"id"`
	Name     string           `json:"name" mapstructure:"name"`
	DbId     int              `json:"db_id" mapstructure:"db_id"`
	Database *domain.Resource `json:"database" mapstructure:"database"`
}

type GroupResource struct {
	Name        string   `json:"name" mapstructure:"name"`
	Permissions []string `json:"permission" mapstructure:"permission"`
	Urn         string   `json:"urn" mapstructure:"urn"`
	Type        string   `json:"type" mapstructure:"type"`
}

type Group struct {
	ID                  int              `json:"id" mapstructure:"id"`
	Name                string           `json:"name" mapstructure:"name"`
	DatabaseResources   []*GroupResource `json:"database" mapstructure:"database"`
	CollectionResources []*GroupResource `json:"collection" mapstructure:"collection"`
}

func (d *Database) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeDatabase {
		return ErrInvalidResourceType
	}

	databaseURN := strings.Split(r.URN, ":")
	if len(databaseURN) != 2 {
		return ErrInvalidDatabaseURN
	}
	id, err := strconv.Atoi(databaseURN[1])
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
		URN:  fmt.Sprintf("database:%v", d.ID),
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

func (t *Table) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeTable {
		return ErrInvalidResourceType
	}

	tableURN := strings.Split(r.URN, ":")
	if len(tableURN) != 2 {
		return ErrInvalidTableURN
	}

	tableURN = strings.Split(tableURN[1], ".")
	id, err := strconv.Atoi(tableURN[1])
	if err != nil {
		return err
	}

	t.ID = id
	t.Name = r.Name
	t.DbId, err = strconv.Atoi(tableURN[0])
	if err != nil {
		return err
	}
	return nil
}

func (t *Table) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type:    ResourceTypeTable,
		Name:    t.Name,
		URN:     fmt.Sprintf("table:%d.%d", t.DbId, t.ID),
		Details: t.Database.Details,
	}
}

func (g *Group) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeGroup {
		return ErrInvalidResourceType
	}

	groupUrn := strings.Split(r.URN, ":")
	if len(groupUrn) != 2 {
		return ErrInvalidGroupURN
	}
	id, err := strconv.Atoi(groupUrn[1])
	if err != nil {
		return err
	}

	g.ID = id
	g.Name = r.Name
	_ = mapstructure.Decode(r.Details["database"], &g.DatabaseResources)
	_ = mapstructure.Decode(r.Details["collection"], &g.CollectionResources)
	if err != nil {
		return err
	}
	return nil
}

func (g *Group) ToDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeGroup,
		Name: g.Name,
		URN:  fmt.Sprintf("group:%d", g.ID),
		Details: map[string]interface{}{
			"database":   g.DatabaseResources,
			"collection": g.CollectionResources,
		},
	}
}

type Collection struct {
	ID              interface{} `json:"id"`
	Name            string      `json:"name"`
	Slug            string      `json:"slug"`
	Location        string      `json:"location,omitempty"`
	Namespace       string      `json:"namespace,omitempty"`
	PersonalOwnerId interface{} `json:"personal_owner_id,omitempty"`
}

func (c *Collection) FromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeCollection {
		return ErrInvalidResourceType
	}

	collectionUrn := strings.Split(r.URN, ":")
	if len(collectionUrn) != 2 {
		return ErrInvalidCollectionURN
	}
	id, err := strconv.Atoi(collectionUrn[1])
	if err != nil {
		return err
	}

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
		URN:     fmt.Sprintf("collection:%v", c.ID),
		Details: details,
	}
}
