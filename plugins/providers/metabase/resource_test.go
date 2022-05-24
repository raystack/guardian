package metabase_test

import (
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/providers/metabase"
	"github.com/stretchr/testify/assert"
)

func TestDatabase(t *testing.T) {
	t.Run("ToDomain", func(t *testing.T) {
		t.Run("should pass right values for type, name, and URN", func(t *testing.T) {
			testCases := []struct {
				db               *metabase.Database
				expectedResource *domain.Resource
			}{
				{
					db: &metabase.Database{
						ID:   1,
						Name: "database 1",
					},
					expectedResource: &domain.Resource{
						Type: metabase.ResourceTypeDatabase,
						Name: "database 1",
						URN:  "database:1",
					},
				},
			}

			for _, tc := range testCases {
				actualResoruce := tc.db.ToDomain()

				assert.Equal(t, tc.expectedResource.Type, actualResoruce.Type)
				assert.Equal(t, tc.expectedResource.Name, actualResoruce.Name)
				assert.Equal(t, tc.expectedResource.URN, actualResoruce.URN)
			}
		})
	})

	t.Run("FromDomain", func(t *testing.T) {
		t.Run("should return error if the resource type is not database", func(t *testing.T) {
			expectedError := metabase.ErrInvalidResourceType

			r := &domain.Resource{
				Type: "not-database-type",
			}
			d := new(metabase.Database)
			actualError := d.FromDomain(r)

			assert.Equal(t, expectedError, actualError)
		})

		t.Run("should return error if got error converting string to int", func(t *testing.T) {
			r := &domain.Resource{
				URN:  "non-numeric",
				Type: metabase.ResourceTypeDatabase,
			}
			d := new(metabase.Database)
			actualError := d.FromDomain(r)

			assert.Error(t, actualError)
		})

		t.Run("should pass right values for id and name", func(t *testing.T) {
			expectedDatabase := &metabase.Database{
				ID:   1,
				Name: "test-resource",
			}

			r := &domain.Resource{
				URN:  "database:1",
				Type: metabase.ResourceTypeDatabase,
				Name: "test-resource",
			}
			d := new(metabase.Database)
			actualError := d.FromDomain(r)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedDatabase, d)
		})
	})
}

func TestCollection(t *testing.T) {
	t.Run("ToDomain", func(t *testing.T) {
		t.Run("should pass right values for type, name, and URN", func(t *testing.T) {
			testCases := []struct {
				db               *metabase.Collection
				expectedResource *domain.Resource
			}{
				{
					db: &metabase.Collection{
						ID:   "root",
						Name: "collection 1",
					},
					expectedResource: &domain.Resource{
						Type: metabase.ResourceTypeCollection,
						Name: "collection 1",
						URN:  "collection:root",
					},
				},
				{
					db: &metabase.Collection{
						ID:   1,
						Name: "collection 2",
					},
					expectedResource: &domain.Resource{
						Type: metabase.ResourceTypeCollection,
						Name: "collection 2",
						URN:  "collection:1",
					},
				},
			}

			for _, tc := range testCases {
				actualResoruce := tc.db.ToDomain()

				assert.Equal(t, tc.expectedResource.Type, actualResoruce.Type)
				assert.Equal(t, tc.expectedResource.Name, actualResoruce.Name)
				assert.Equal(t, tc.expectedResource.URN, actualResoruce.URN)
			}
		})
	})

	t.Run("FromDomain", func(t *testing.T) {
		t.Run("should return error if the resource type is not collection", func(t *testing.T) {
			expectedErr := metabase.ErrInvalidResourceType

			r := &domain.Resource{
				Type: "not-collection-type",
			}

			c := new(metabase.Collection)
			actualError := c.FromDomain(r)

			assert.Equal(t, expectedErr, actualError)
		})

		t.Run("should pass right values for id", func(t *testing.T) {
			testCases := []struct {
				expectedResource   *domain.Resource
				expectedCollection *metabase.Collection
			}{
				{
					expectedResource: &domain.Resource{
						URN:  "collection:2",
						Name: "test-collection",
						Type: metabase.ResourceTypeCollection,
					},
					expectedCollection: &metabase.Collection{
						ID:   2,
						Name: "test-collection",
					},
				},
				{
					expectedResource: &domain.Resource{
						URN:  "collection:1",
						Name: "test-collection",
						Type: metabase.ResourceTypeCollection,
					},
					expectedCollection: &metabase.Collection{
						ID:   1,
						Name: "test-collection",
					},
				},
			}

			for _, tc := range testCases {
				c := new(metabase.Collection)

				actualError := c.FromDomain(tc.expectedResource)

				assert.Nil(t, actualError)
				assert.Equal(t, tc.expectedCollection, c)
			}
		})
	})
}
