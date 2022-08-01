package grafana_test

import (
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/providers/grafana"
	"github.com/stretchr/testify/assert"
)

func TestDashboard(t *testing.T) {
	t.Run("ToDomain", func(t *testing.T) {
		t.Run("should pass right values for type, name, and URN", func(t *testing.T) {
			testCases := []struct {
				db               *grafana.Dashboard
				expectedResource *domain.Resource
			}{
				{
					db: &grafana.Dashboard{
						ID:          1,
						Title:       "dashboard 1",
						FolderID:    101,
						FolderUID:   "Test-folder-UID",
						FolderTitle: "Title",
						UID:         "Test-UID",
					},
					expectedResource: &domain.Resource{
						Type: grafana.ResourceTypeDashboard,
						Name: "dashboard 1",
						URN:  "1",
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
		t.Run("should return error if the resource type is not dashboard", func(t *testing.T) {
			expectedError := grafana.ErrInvalidResourceType

			r := &domain.Resource{
				Type: "not-dashboard-type",
			}
			d := new(grafana.Dashboard)
			actualError := d.FromDomain(r)

			assert.Equal(t, expectedError, actualError)
		})

		t.Run("should return error if got error converting string to int", func(t *testing.T) {
			r := &domain.Resource{
				URN:  "non-numeric",
				Type: grafana.ResourceTypeDashboard,
			}
			d := new(grafana.Dashboard)
			actualError := d.FromDomain(r)

			assert.Error(t, actualError)
		})

		t.Run("should pass right values for id and name", func(t *testing.T) {
			expectedDashboard := &grafana.Dashboard{
				ID:    1,
				Title: "test-resource",
			}

			r := &domain.Resource{
				URN:  "1",
				Type: grafana.ResourceTypeDashboard,
				Name: "test-resource",
			}
			d := new(grafana.Dashboard)
			actualError := d.FromDomain(r)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedDashboard, d)
		})
	})
}
