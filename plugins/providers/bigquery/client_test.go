package bigquery_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goto/guardian/plugins/providers/bigquery"
	"github.com/stretchr/testify/suite"
	bqApi "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/option"
)

type ClientTestSuite struct {
	suite.Suite
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (s *ClientTestSuite) TestGetDatasets() {
	projectID := "test_project"
	expectedDatasets := []*bigquery.Dataset{
		{
			ProjectID: projectID,
			DatasetID: "dataset_1",
			Labels: map[string]string{
				"test_label": "test_value",
			},
		},
		{
			ProjectID: projectID,
			DatasetID: "dataset_2",
		},
	}

	isFirstPageHelper := true
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := &bqApi.DatasetList{}
		if isFirstPageHelper {
			res.NextPageToken = "test_next_page_token"
			res.Datasets = []*bqApi.DatasetListDatasets{
				{
					Id: projectID + ":dataset_1",
					DatasetReference: &bqApi.DatasetReference{
						DatasetId: "dataset_1",
						ProjectId: projectID,
					},
					Labels: map[string]string{
						"test_label": "test_value",
					},
				},
			}
			isFirstPageHelper = false
		} else { // 2nd page
			res.Datasets = []*bqApi.DatasetListDatasets{
				{
					Id: projectID + ":dataset_2",
					DatasetReference: &bqApi.DatasetReference{
						DatasetId: "dataset_2",
						ProjectId: projectID,
					},
				},
			}
		}

		b, err := json.Marshal(res)
		if err != nil {
			http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(b)
	}))
	defer ts.Close()

	client, err := bigquery.NewBigQueryClient(projectID, option.WithoutAuthentication(), option.WithEndpoint(ts.URL))
	s.Require().NoError(err)

	datasets, err := client.GetDatasets(context.Background())

	s.Nil(err)
	s.Equal(expectedDatasets, datasets)
}

func (s *ClientTestSuite) TestGetTables() {
	projectID := "test_project"
	datasetID := "test_dataset"
	expectedTables := []*bigquery.Table{
		{
			ProjectID: projectID,
			DatasetID: datasetID,
			TableID:   "table_1",
			Labels: map[string]string{
				"test_label": "test_value",
			},
		},
		{
			ProjectID: projectID,
			DatasetID: datasetID,
			TableID:   "table_2",
		},
	}

	isFirstPageHelper := true
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := &bqApi.TableList{}
		if isFirstPageHelper {
			res.NextPageToken = "test_next_page_token"
			res.Tables = []*bqApi.TableListTables{
				{
					Id: projectID + ":" + datasetID + ".table_1",
					TableReference: &bqApi.TableReference{
						ProjectId: projectID,
						DatasetId: datasetID,
						TableId:   "table_1",
					},
					Labels: map[string]string{
						"test_label": "test_value",
					},
				},
			}
			isFirstPageHelper = false
		} else { // 2nd page
			res.Tables = []*bqApi.TableListTables{
				{
					Id: projectID + ":" + datasetID + ".table_2",
					TableReference: &bqApi.TableReference{
						ProjectId: projectID,
						DatasetId: datasetID,
						TableId:   "table_2",
					},
				},
			}
		}

		b, err := json.Marshal(res)
		if err != nil {
			http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(b)
	}))
	defer ts.Close()

	client, err := bigquery.NewBigQueryClient(projectID, option.WithoutAuthentication(), option.WithEndpoint(ts.URL))
	s.Require().NoError(err)

	tables, err := client.GetTables(context.Background(), datasetID)

	s.Nil(err)
	s.Equal(expectedTables, tables)
}
