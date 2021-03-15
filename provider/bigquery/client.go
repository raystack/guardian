package bigquery

import (
	"context"

	bq "cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client is wrapper for bigquery client
type Client struct {
	client *bq.Client
}

// NewClient returns *bigquery.Client
func NewClient(projectID string, credentialsJSON []byte) (*Client, error) {
	ctx := context.Background()
	client, err := bq.NewClient(ctx, projectID, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, err
	}

	return &Client{client}, nil
}

// GetDatasets returns all datasets within a project
func (c *Client) GetDatasets(ctx context.Context) ([]*Dataset, error) {
	var results []*Dataset
	it := c.client.Datasets(ctx)
	for {
		dataset, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		results = append(results, &Dataset{
			ProjectID: dataset.ProjectID,
			DatasetID: dataset.DatasetID,
		})
	}

	return results, nil
}

// GetTables returns all tables within a dataset
func (c *Client) GetTables(ctx context.Context, datasetID string) ([]*Table, error) {
	var results []*Table
	it := c.client.Dataset(datasetID).Tables(ctx)
	for {
		table, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		results = append(results, &Table{
			ProjectID: table.ProjectID,
			DatasetID: table.DatasetID,
			TableID:   table.TableID,
		})
	}

	return results, nil
}
