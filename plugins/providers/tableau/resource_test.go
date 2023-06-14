package tableau_test

import (
	"testing"

	"github.com/raystack/guardian/domain"
	"github.com/raystack/guardian/plugins/providers/tableau"
	"github.com/stretchr/testify/assert"
)

func TestWorkbook(t *testing.T) {
	t.Run("ToDomain", func(t *testing.T) {
		t.Run("should pass right values for type, name, and URN", func(t *testing.T) {
			testCases := []struct {
				wb               *tableau.Workbook
				expectedResource *domain.Resource
			}{
				{
					wb: &tableau.Workbook{
						ID:   "1",
						Name: "workbook 1",
					},
					expectedResource: &domain.Resource{
						Type: tableau.ResourceTypeWorkbook,
						Name: "workbook 1",
						URN:  "1",
					},
				},
			}

			for _, tc := range testCases {
				actualResoruce := tc.wb.ToDomain()

				assert.Equal(t, tc.expectedResource.Type, actualResoruce.Type)
				assert.Equal(t, tc.expectedResource.Name, actualResoruce.Name)
				assert.Equal(t, tc.expectedResource.URN, actualResoruce.URN)
			}
		})
	})

	t.Run("FromDomain", func(t *testing.T) {
		t.Run("should return error if the resource type is not workbook", func(t *testing.T) {
			expectedError := tableau.ErrInvalidResourceType

			r := &domain.Resource{
				Type: "not-workbook-type",
			}
			d := new(tableau.Workbook)
			actualError := d.FromDomain(r)

			assert.Equal(t, expectedError, actualError)
		})

		t.Run("should pass right values for id and name", func(t *testing.T) {
			expectedWorkbook := &tableau.Workbook{
				ID:   "1",
				Name: "test-resource",
			}

			r := &domain.Resource{
				URN:  "1",
				Type: tableau.ResourceTypeWorkbook,
				Name: "test-resource",
			}
			d := new(tableau.Workbook)
			actualError := d.FromDomain(r)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedWorkbook, d)
		})
	})
}

func TestFlow(t *testing.T) {
	t.Run("ToDomain", func(t *testing.T) {
		t.Run("should pass right values for type, name, and URN", func(t *testing.T) {
			testCases := []struct {
				fl               *tableau.Flow
				expectedResource *domain.Resource
			}{
				{
					fl: &tableau.Flow{
						ID:   "1",
						Name: "flow 1",
					},
					expectedResource: &domain.Resource{
						Type: tableau.ResourceTypeFlow,
						Name: "flow 1",
						URN:  "1",
					},
				},
			}

			for _, tc := range testCases {
				actualResoruce := tc.fl.ToDomain()

				assert.Equal(t, tc.expectedResource.Type, actualResoruce.Type)
				assert.Equal(t, tc.expectedResource.Name, actualResoruce.Name)
				assert.Equal(t, tc.expectedResource.URN, actualResoruce.URN)
			}
		})
	})

	t.Run("FromDomain", func(t *testing.T) {
		t.Run("should return error if the resource type is not flow", func(t *testing.T) {
			expectedError := tableau.ErrInvalidResourceType

			r := &domain.Resource{
				Type: "not-flow-type",
			}
			f := new(tableau.Flow)
			actualError := f.FromDomain(r)

			assert.Equal(t, expectedError, actualError)
		})

		t.Run("should pass right values for id and name", func(t *testing.T) {
			expectedFlow := &tableau.Flow{
				ID:   "1",
				Name: "test-resource",
			}

			r := &domain.Resource{
				URN:  "1",
				Type: tableau.ResourceTypeFlow,
				Name: "test-resource",
			}
			f := new(tableau.Flow)
			actualError := f.FromDomain(r)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedFlow, f)
		})
	})
}

func TestDatasource(t *testing.T) {
	t.Run("ToDomain", func(t *testing.T) {
		t.Run("should pass right values for type, name, and URN", func(t *testing.T) {
			testCases := []struct {
				ds               *tableau.DataSource
				expectedResource *domain.Resource
			}{
				{
					ds: &tableau.DataSource{
						ID:   "1",
						Name: "datasource 1",
					},
					expectedResource: &domain.Resource{
						Type: tableau.ResourceTypeDataSource,
						Name: "datasource 1",
						URN:  "1",
					},
				},
			}

			for _, tc := range testCases {
				actualResoruce := tc.ds.ToDomain()

				assert.Equal(t, tc.expectedResource.Type, actualResoruce.Type)
				assert.Equal(t, tc.expectedResource.Name, actualResoruce.Name)
				assert.Equal(t, tc.expectedResource.URN, actualResoruce.URN)
			}
		})
	})

	t.Run("FromDomain", func(t *testing.T) {
		t.Run("should return error if the resource type is not datasource", func(t *testing.T) {
			expectedError := tableau.ErrInvalidResourceType

			r := &domain.Resource{
				Type: "not-datasource-type",
			}
			d := new(tableau.DataSource)
			actualError := d.FromDomain(r)

			assert.Equal(t, expectedError, actualError)
		})

		t.Run("should pass right values for id and name", func(t *testing.T) {
			expectedDatasource := &tableau.DataSource{
				ID:   "1",
				Name: "test-resource",
			}

			r := &domain.Resource{
				URN:  "1",
				Type: tableau.ResourceTypeDataSource,
				Name: "test-resource",
			}
			d := new(tableau.DataSource)
			actualError := d.FromDomain(r)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedDatasource, d)
		})
	})
}

func TestView(t *testing.T) {
	t.Run("ToDomain", func(t *testing.T) {
		t.Run("should pass right values for type, name, and URN", func(t *testing.T) {
			testCases := []struct {
				vw               *tableau.View
				expectedResource *domain.Resource
			}{
				{
					vw: &tableau.View{
						ID:   "1",
						Name: "view 1",
					},
					expectedResource: &domain.Resource{
						Type: tableau.ResourceTypeView,
						Name: "view 1",
						URN:  "1",
					},
				},
			}

			for _, tc := range testCases {
				actualResoruce := tc.vw.ToDomain()

				assert.Equal(t, tc.expectedResource.Type, actualResoruce.Type)
				assert.Equal(t, tc.expectedResource.Name, actualResoruce.Name)
				assert.Equal(t, tc.expectedResource.URN, actualResoruce.URN)
			}
		})
	})

	t.Run("FromDomain", func(t *testing.T) {
		t.Run("should return error if the resource type is not view", func(t *testing.T) {
			expectedError := tableau.ErrInvalidResourceType

			r := &domain.Resource{
				Type: "not-view-type",
			}
			v := new(tableau.View)
			actualError := v.FromDomain(r)

			assert.Equal(t, expectedError, actualError)
		})

		t.Run("should pass right values for id and name", func(t *testing.T) {
			expectedView := &tableau.View{
				ID:   "1",
				Name: "test-resource",
			}

			r := &domain.Resource{
				URN:  "1",
				Type: tableau.ResourceTypeView,
				Name: "test-resource",
			}
			v := new(tableau.View)
			actualError := v.FromDomain(r)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedView, v)
		})
	})
}

func TestMetric(t *testing.T) {
	t.Run("ToDomain", func(t *testing.T) {
		t.Run("should pass right values for type, name, and URN", func(t *testing.T) {
			testCases := []struct {
				mt               *tableau.Metric
				expectedResource *domain.Resource
			}{
				{
					mt: &tableau.Metric{
						ID:   "1",
						Name: "metric 1",
					},
					expectedResource: &domain.Resource{
						Type: tableau.ResourceTypeMetric,
						Name: "metric 1",
						URN:  "1",
					},
				},
			}

			for _, tc := range testCases {
				actualResoruce := tc.mt.ToDomain()

				assert.Equal(t, tc.expectedResource.Type, actualResoruce.Type)
				assert.Equal(t, tc.expectedResource.Name, actualResoruce.Name)
				assert.Equal(t, tc.expectedResource.URN, actualResoruce.URN)
			}
		})
	})

	t.Run("FromDomain", func(t *testing.T) {
		t.Run("should return error if the resource type is not metric", func(t *testing.T) {
			expectedError := tableau.ErrInvalidResourceType

			r := &domain.Resource{
				Type: "not-metric-type",
			}
			m := new(tableau.Metric)
			actualError := m.FromDomain(r)

			assert.Equal(t, expectedError, actualError)
		})

		t.Run("should pass right values for id and name", func(t *testing.T) {
			expectedMetric := &tableau.Metric{
				ID:   "1",
				Name: "test-resource",
			}

			r := &domain.Resource{
				URN:  "1",
				Type: tableau.ResourceTypeMetric,
				Name: "test-resource",
			}
			m := new(tableau.Metric)
			actualError := m.FromDomain(r)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedMetric, m)
		})
	})
}
