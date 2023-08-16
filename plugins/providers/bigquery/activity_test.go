package bigquery_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/plugins/providers/bigquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/logging/v2"
	"google.golang.org/genproto/googleapis/cloud/audit"
)

func TestActivity_ToProviderActivity(t *testing.T) {
	now := time.Now()
	dummyProvider := domain.Provider{
		ID:   "dummy-provider-id",
		Type: "bigquery",
		URN:  "dummy-provider-urn",
	}
	auditLog := &audit.AuditLog{
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
	}
	auditLogBytes, err := json.Marshal(auditLog)
	if err != nil {
		require.NoError(t, err)
	}
	a := bigquery.Activity{
		&logging.LogEntry{
			InsertId:     "dummy-insert-id",
			Timestamp:    now.Format(time.RFC3339Nano),
			ProtoPayload: googleapi.RawMessage(auditLogBytes),
		},
	}

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
			"insert_id":       a.InsertId,
			"severity":        "",
			"resource":        nil,
			"labels":          nil,
			"operation":       nil,
			"trace":           "",
			"source_location": nil,
			"timestamp":       a.Timestamp,
			"span_id":         "",
			"trace_sampled":   false,
		},
	}

	expectedTimestamp, err := time.Parse(time.RFC3339Nano, a.Timestamp)
	require.NoError(t, err)

	expectedActivity := &domain.Activity{
		ProviderID:         dummyProvider.ID,
		ProviderActivityID: "dummy-insert-id",
		Timestamp:          expectedTimestamp,
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
