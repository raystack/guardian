package bigquery_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/logging"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/providers/bigquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/cloud/audit"
)

func TestActivity_ToActivity(t *testing.T) {
	now := time.Now()
	dummyProvider := domain.Provider{
		ID:   "dummy-provider-id",
		Type: "bigquery",
		URN:  "dummy-provider-urn",
	}
	a := bigquery.Activity{
		&logging.Entry{
			Timestamp: now,
			Payload: &audit.AuditLog{
				MethodName: "test-method-name",
				AuthenticationInfo: &audit.AuthenticationInfo{
					PrincipalEmail: "test-principal-email",
				},
				AuthorizationInfo: []*audit.AuthorizationInfo{
					{
						Permission: "test-permission",
					},
				},
				ResourceName: "projects/xxx/datasets/yyy/tables/zzz",
			},
		},
	}

	expectedMetadata := map[string]interface{}{
		"method_name": "test-method-name",
		"authentication_info": map[string]interface{}{
			"principal_email": "test-principal-email",
		},
		"authorization_info": []interface{}{
			map[string]interface{}{
				"permission": "test-permission",
			},
		},
		"resource_name": "projects/xxx/datasets/yyy/tables/zzz",
	}

	expectedActivity := &domain.Activity{
		ProviderID:     dummyProvider.ID,
		Timestamp:      a.Timestamp,
		Type:           "test-method-name",
		AccountID:      "test-principal-email",
		AccountType:    "user",
		Metadata:       expectedMetadata,
		Authorizations: []string{"test-permission"},
		Resource: &domain.Resource{
			ProviderType: "bigquery",
			ProviderURN:  "dummy-provider-urn",
			Type:         bigquery.ResourceTypeTable,
			URN:          "xxx:yyy.zzz",
		},
	}

	actualActivity, err := a.ToActivity(dummyProvider)

	assert.NoError(t, err)
	assert.Equal(t, expectedActivity, actualActivity)
}

func TestList(t *testing.T) {
	c, err := bigquery.NewCloudLoggingClient(context.Background(), "pilotdata-integration", nil)
	require.NoError(t, err)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	anHourAgo := time.Now().Add(-1 * time.Hour)

	f := domain.ImportActivitiesFilter{
		ResourceIDs: []string{"1", "2"},
	}
	err = f.PopulateResources(map[string]*domain.Resource{
		"1": {
			Type: "table",
			URN:  "pilotdata-integration:playground.name_event",
		},
		"2": {
			Type: "dataset",
			URN:  "pilotdata-integration:playground",
		},
	})
	require.NoError(t, err)
	es, err := c.ListLogEntries(ctx, bigquery.ImportActivitiesFilter{
		ImportActivitiesFilter: domain.ImportActivitiesFilter{
			ResourceIDs:  []string{"1", "2"},
			TimestampGte: &anHourAgo,
		},
		Limit: 5,
	})
	require.NoError(t, err)

	for _, e := range es {
		fmt.Printf("e: %+v\n", e)
		fmt.Printf("e.Resource: %+v\n", e.Resource)
	}
	t.Fatal("done")
}
