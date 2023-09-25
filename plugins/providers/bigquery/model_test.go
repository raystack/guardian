package bigquery_test

import (
	"testing"

	"github.com/raystack/guardian/core/resource"

	"github.com/raystack/guardian/domain"
	"github.com/raystack/guardian/plugins/providers/bigquery"
	"github.com/stretchr/testify/assert"
)

func TestDataSet(t *testing.T) {
	t.Run("ToDomain", func(t *testing.T) {
		t.Run("should pass right values for type, name, and URN", func(t *testing.T) {
			testCases := []struct {
				ds               *bigquery.Dataset
				expectedResource *domain.Resource
			}{
				{
					ds: &bigquery.Dataset{
						ProjectID: "p_id",
						DatasetID: "d_id",
					},
					expectedResource: &domain.Resource{
						Type: bigquery.ResourceTypeDataset,
						Name: "d_id",
						URN:  "p_id:d_id",
					},
				},
				{
					ds: &bigquery.Dataset{
						ProjectID: "p_id",
						DatasetID: "d_id",
						Labels:    nil,
					},
					expectedResource: &domain.Resource{
						Type: bigquery.ResourceTypeDataset,
						Name: "d_id",
						URN:  "p_id:d_id",
					},
				},
				{
					ds: &bigquery.Dataset{
						ProjectID: "p_id",
						DatasetID: "d_id",
						Labels: map[string]string{
							"key1": "value1",
						},
					},
					expectedResource: &domain.Resource{
						Type: bigquery.ResourceTypeDataset,
						Name: "d_id",
						URN:  "p_id:d_id",
						Details: map[string]interface{}{
							resource.ReservedDetailsKeyMetadata: map[string]interface{}{
								"labels": map[string]string{
									"key1": "value1",
								},
							},
						},
					},
				},
			}

			for _, tc := range testCases {
				actualResource := tc.ds.ToDomain()

				assert.Equal(t, tc.expectedResource.Type, actualResource.Type)
				assert.Equal(t, tc.expectedResource.Name, actualResource.Name)
				assert.Equal(t, tc.expectedResource.URN, actualResource.URN)
				assert.Equal(t, tc.expectedResource.Details, actualResource.Details)
			}
		})
	})

	t.Run("FromDomain", func(t *testing.T) {
		t.Run("should return error if return type is not dataset", func(t *testing.T) {
			expectedError := bigquery.ErrInvalidResourceType
			r := &domain.Resource{
				Type: "not-dataset-type",
			}
			d := new(bigquery.Dataset)

			actualError := d.FromDomain(r)

			assert.Equal(t, expectedError, actualError)
		})

		t.Run("should return no error when ProjectID and DatasetID are correct", func(t *testing.T) {
			expectedDataset := &bigquery.Dataset{
				ProjectID: "p_id",
				DatasetID: "d_id",
			}
			r := &domain.Resource{
				Type: bigquery.ResourceTypeDataset,
				Name: "d_id",
				URN:  "p_id",
			}
			d := new(bigquery.Dataset)
			actualError := d.FromDomain(r)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedDataset, d)
		})
	})
}

func TestTable(t *testing.T) {
	t.Run("ToDomain", func(t *testing.T) {
		t.Run("should pass right values for type, name, and URN", func(t *testing.T) {
			testCases := []struct {
				tb               *bigquery.Table
				expectedResource *domain.Resource
			}{
				{
					tb: &bigquery.Table{
						TableID:   "t_id",
						ProjectID: "p_id",
						DatasetID: "d_id",
					},
					expectedResource: &domain.Resource{
						Type: bigquery.ResourceTypeTable,
						Name: "t_id",
						URN:  "p_id:d_id.t_id",
					},
				},
				{
					tb: &bigquery.Table{
						TableID:   "t_id",
						ProjectID: "p_id",
						DatasetID: "d_id",
						Labels:    nil,
					},
					expectedResource: &domain.Resource{
						Type: bigquery.ResourceTypeTable,
						Name: "t_id",
						URN:  "p_id:d_id.t_id",
					},
				},
				{
					tb: &bigquery.Table{
						TableID:   "t_id",
						ProjectID: "p_id",
						DatasetID: "d_id",
						Labels: map[string]string{
							"key1": "value1",
						},
					},
					expectedResource: &domain.Resource{
						Type: bigquery.ResourceTypeTable,
						Name: "t_id",
						URN:  "p_id:d_id.t_id",
						Details: map[string]interface{}{
							resource.ReservedDetailsKeyMetadata: map[string]interface{}{
								"labels": map[string]string{
									"key1": "value1",
								},
							},
						},
					},
				},
			}

			for _, tc := range testCases {
				actualResource := tc.tb.ToDomain()

				assert.Equal(t, tc.expectedResource.Type, actualResource.Type)
				assert.Equal(t, tc.expectedResource.Name, actualResource.Name)
				assert.Equal(t, tc.expectedResource.URN, actualResource.URN)
				assert.Equal(t, tc.expectedResource.Details, actualResource.Details)
			}
		})
	})

	t.Run("FromDomain", func(t *testing.T) {
		t.Run("should return error if resource type is not table", func(t *testing.T) {
			expectedError := bigquery.ErrInvalidResourceType
			r := &domain.Resource{
				Type: "not-table-type",
			}
			d := new(bigquery.Table)

			actualError := d.FromDomain(r)

			assert.Equal(t, expectedError, actualError)
		})

		t.Run("should return error invalidTableURN", func(t *testing.T) {
			expectedError := bigquery.ErrInvalidTableURN
			r := &domain.Resource{
				Type: bigquery.ResourceTypeTable,
				URN:  "wrong_table_URN",
			}
			d := new(bigquery.Table)

			actualError := d.FromDomain(r)
			assert.Equal(t, expectedError, actualError)
		})

		t.Run("should return no error", func(t *testing.T) {
			expectedTable := &bigquery.Table{
				TableID:   "t_id",
				ProjectID: "p_id",
				DatasetID: "d_id",
			}
			r := &domain.Resource{
				Type: bigquery.ResourceTypeTable,
				Name: "t_id",
				URN:  "p_id:d_id.t_id",
			}
			d := new(bigquery.Table)
			actualError := d.FromDomain(r)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedTable, d)
		})
	})
}

func TestBigQueryResourceName(t *testing.T) {
	testCases := []struct {
		bqrn               bigquery.BigQueryResourceName
		expectedProjectID  string
		expectedDatasetID  string
		expectedTableID    string
		expectedResourceID string
	}{
		{
			bqrn:               "projects/p_id/datasets/d_id/tables/t_id",
			expectedProjectID:  "p_id",
			expectedDatasetID:  "d_id",
			expectedTableID:    "t_id",
			expectedResourceID: "p_id:d_id.t_id",
		},
		{
			bqrn:               "projects/p_id/datasets/d_id",
			expectedProjectID:  "p_id",
			expectedDatasetID:  "d_id",
			expectedTableID:    "",
			expectedResourceID: "p_id:d_id",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedProjectID, tc.bqrn.ProjectID())
		assert.Equal(t, tc.expectedDatasetID, tc.bqrn.DatasetID())
		assert.Equal(t, tc.expectedTableID, tc.bqrn.TableID())
		assert.Equal(t, tc.expectedResourceID, tc.bqrn.BigQueryResourceID())
	}
}
