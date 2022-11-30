package bigquery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/logadmin"
	"github.com/odpf/guardian/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/cloud/audit"
)

var (
	ErrInvalidActivityPayloadType = errors.New("payload is not of type *audit.AuditLog")
	ErrEmptyActivityPayload       = errors.New("couldn't get payload from log entry")
)

type auditLog struct {
	*audit.AuditLog
}

func (a auditLog) GetAccountType() string {
	if a.AuthenticationInfo.ServiceAccountKeyName != "" {
		return AccountTypeServiceAccount
	}
	return AccountTypeUser
}

func (a auditLog) GetResource(p domain.Provider) *domain.Resource {
	rn := BigQueryResourceName(a.ResourceName)
	resourceType := ResourceTypeDataset
	if tableID := rn.TableID(); tableID != "" {
		resourceType = ResourceTypeTable
	}

	return &domain.Resource{
		ProviderType: p.Type,
		ProviderURN:  p.URN,
		Type:         resourceType,
		URN:          rn.BigQueryResourceID(),
	}
}

type Activity struct {
	*logging.Entry
}

func (a Activity) getAuditLog() (*auditLog, error) {
	l, ok := a.Payload.(*audit.AuditLog)
	if !ok {
		return nil, fmt.Errorf("%w: %T", ErrInvalidActivityPayloadType, a.Payload)
	}
	return &auditLog{l}, nil
}

func (a Activity) ToDomainActivity(p domain.Provider) (*domain.Activity, error) {
	activity := &domain.Activity{
		ProviderID:         p.ID,
		Timestamp:          a.Timestamp,
		ProviderActivityID: a.InsertID,
	}

	al, err := a.getAuditLog()
	if err != nil {
		return nil, err
	}
	if al == nil {
		return nil, ErrEmptyActivityPayload
	}

	activity.Type = al.MethodName
	if al.AuthenticationInfo != nil {
		activity.AccountType = al.GetAccountType()
		activity.AccountID = al.AuthenticationInfo.PrincipalEmail
	}

	loggingEntryMetadata := map[string]interface{}{}
	loggingEntryMap := map[string]interface{}{
		"payload":         al,
		"insert_id":       a.InsertID,
		"severity":        a.Severity,
		"resource":        a.Resource,
		"labels":          a.Labels,
		"operation":       a.Operation,
		"trace":           a.Trace,
		"source_location": a.SourceLocation,
		"timestamp":       a.Timestamp,
		"span_id":         a.SpanID,
		"trace_sampled":   a.TraceSampled,
	}
	if jsonData, err := json.Marshal(loggingEntryMap); err != nil {
		return nil, fmt.Errorf("marshalling payload: %w", err)
	} else if err := json.Unmarshal(jsonData, &loggingEntryMetadata); err != nil {
		return nil, fmt.Errorf("unmarshalling payload to metadata: %w", err)
	}

	if activity.Metadata == nil {
		activity.Metadata = map[string]interface{}{}
	}
	activity.Metadata["logging_entry"] = loggingEntryMetadata

	for _, ai := range al.AuthorizationInfo {
		activity.Authorizations = append(activity.Authorizations, ai.Permission)
	}

	activity.Resource = al.GetResource(p)

	return activity, nil
}

type cloudLoggingClient struct {
	client *logadmin.Client
}

func NewCloudLoggingClient(ctx context.Context, projectID string, credentialsJSON []byte) (*cloudLoggingClient, error) {
	var options []option.ClientOption
	if credentialsJSON != nil {
		options = append(options, option.WithCredentialsJSON(credentialsJSON))
	}
	client, err := logadmin.NewClient(ctx, projectID, options...)
	if err != nil {
		return nil, err
	}

	return &cloudLoggingClient{
		client: client,
	}, nil
}

type bqResource domain.Resource

func (r bqResource) fullURN() string {
	urn := strings.Split(r.URN, ":")
	if len(urn) < 2 {
		return ""
	}

	projectID := urn[0]
	s := fmt.Sprintf(`projects/%s/datasets/%s`, projectID, urn[1])

	if r.Type == ResourceTypeTable {
		urn := strings.Split(urn[1], ".")
		if len(urn) < 2 {
			return ""
		}
		s = fmt.Sprintf(`projects/%s/datasets/%s/tables/%s`, projectID, urn[0], urn[1])
	}

	return s
}

type ImportActivitiesFilter struct {
	domain.ImportActivitiesFilter
	Types          []string
	Authorizations []string
	Limit          int
}

type bqFilter ImportActivitiesFilter

func (f bqFilter) String() string {
	criterias := []string{
		`protoPayload.serviceName="bigquery.googleapis.com"`,
		`resource.type="bigquery_dataset"`, // exclude logs for bigquery jobs ("bigquery_project")
	}

	if len(f.AccountIDs) > 0 {
		criterias = append(criterias,
			fmt.Sprintf(`protoPayload.authenticationInfo.principalEmail=("%s")`, strings.Join(f.AccountIDs, `" OR "`)),
		)
	}

	resources := f.GetResources()
	if len(resources) > 0 {
		resourceNames := []string{}
		for _, r := range resources {
			resourceNames = append(resourceNames, (*bqResource)(r).fullURN())
		}
		criterias = append(criterias, fmt.Sprintf(`protoPayload.resourceName:("%s")`, strings.Join(resourceNames, `" OR "`)))
		// uses ":" (has/contains) operator instead of "=" (equals) operator for resource name so that the result will also
		// include activities on tables under the specified dataset (e.g. "projects/xxx/datasets/yyy" will also include
		// activities on "projects/xxx/datasets/yyy/tables/zzz")
	}
	if len(f.Types) > 0 {
		criterias = append(criterias, fmt.Sprintf(`protoPayload.methodName=("%s")`, strings.Join(f.Types, `" OR "`)))
	}
	// TODO: authorizations
	if f.TimestampGte != nil {
		criterias = append(criterias, fmt.Sprintf(`timestamp>="%s"`, f.TimestampGte.Format(time.RFC3339)))
	}
	if f.TimestampLte != nil {
		criterias = append(criterias, fmt.Sprintf(`timestamp<="%s"`, f.TimestampLte.Format(time.RFC3339)))
	}

	return strings.Join(criterias, " AND ")
}

func (c *cloudLoggingClient) ListLogEntries(ctx context.Context, filter ImportActivitiesFilter) ([]*Activity, error) {
	var entries []*Activity

	options := []logadmin.EntriesOption{logadmin.Filter(bqFilter(filter).String())}
	if filter.Limit > 0 {
		options = append(options, logadmin.NewestFirst())
	}
	it := c.client.Entries(ctx, options...)
	for {
		if filter.Limit > 0 && len(entries) >= filter.Limit {
			break
		}

		e, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, fmt.Errorf("iterating over cloud logging entries: %w", err)
		}
		if e != nil {
			entries = append(entries, &Activity{e})
		}
	}

	return entries, nil
}

func (c *cloudLoggingClient) Close() error {
	return c.client.Close()
}
