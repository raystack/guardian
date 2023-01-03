package bigquery_test

import (
	"testing"
	"time"

	"cloud.google.com/go/logging"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/providers/bigquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/cloud/audit"
)

func TestActivity_ToProviderActivity(t *testing.T) {
	now := time.Now()
	dummyProvider := domain.Provider{
		ID:   "dummy-provider-id",
		Type: "bigquery",
		URN:  "dummy-provider-urn",
	}
	a := bigquery.Activity{
		&logging.Entry{
			InsertID:  "dummy-insert-id",
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

	timestampStr, err := a.Timestamp.MarshalJSON()
	require.NoError(t, err)
	timestampStr = timestampStr[1 : len(timestampStr)-1] // trim quotes
	expectedMetadata := map[string]interface{}{
		"logging_entry": map[string]interface{}{
			"payload": map[string]interface{}{
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
			},
			"insert_id":       a.InsertID,
			"severity":        float64(0),
			"resource":        nil,
			"labels":          nil,
			"operation":       nil,
			"trace":           "",
			"source_location": nil,
			"timestamp":       string(timestampStr),
			"span_id":         "",
			"trace_sampled":   false,
		},
	}

	expectedActivity := &domain.Activity{
		ProviderID:         dummyProvider.ID,
		ProviderActivityID: "dummy-insert-id",
		Timestamp:          a.Timestamp,
		Type:               "test-method-name",
		AccountID:          "test-principal-email",
		AccountType:        "user",
		Metadata:           expectedMetadata,
		Authorizations:     []string{"test-permission"},
		Resource: &domain.Resource{
			ProviderType: "bigquery",
			ProviderURN:  "dummy-provider-urn",
			Type:         bigquery.ResourceTypeTable,
			URN:          "xxx:yyy.zzz",
			Name:         "zzz",
		},
	}

	actualActivity, err := a.ToDomainActivity(dummyProvider)

	assert.NoError(t, err)
	assert.Equal(t, expectedActivity, actualActivity)
}
