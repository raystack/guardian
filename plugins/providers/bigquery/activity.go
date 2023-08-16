package bigquery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goto/guardian/domain"
	"google.golang.org/api/logging/v2"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/cloud/audit"
)

var (
	ErrInvalidActivityPayloadType = errors.New("payload is not of type *audit.AuditLog")
	ErrEmptyActivityPayload       = errors.New("couldn't get payload from log entry")
)

const (
	PrivateLogViewerPermission = "logging.privateLogEntries.list"
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
	name := rn.DatasetID()
	if tableID := rn.TableID(); tableID != "" {
		resourceType = ResourceTypeTable
		name = tableID
	}

	return &domain.Resource{
		ProviderType: p.Type,
		ProviderURN:  p.URN,
		Type:         resourceType,
		URN:          rn.BigQueryResourceID(),
		Name:         name,
	}
}

type Activity struct {
	*logging.LogEntry
}

func (a Activity) getAuditLog() (*auditLog, error) {
	payload, err := a.ProtoPayload.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshalling proto payload: %w", err)
	}
	var al audit.AuditLog
	if err := json.Unmarshal(payload, &al); err != nil {
		return nil, fmt.Errorf("unmarshalling proto payload: %w", err)
	}
	return &auditLog{&al}, nil
}

func (a Activity) ToDomainActivity(p domain.Provider) (*domain.Activity, error) {
	t, err := time.Parse(time.RFC3339Nano, a.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("parsing timestamp: %w", err)
	}

	activity := &domain.Activity{
		ProviderID:         p.ID,
		Timestamp:          t,
		ProviderActivityID: a.InsertId,
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
		"insert_id":       a.InsertId,
		"severity":        a.Severity,
		"resource":        a.Resource,
		"labels":          a.Labels,
		"operation":       a.Operation,
		"trace":           a.Trace,
		"source_location": a.SourceLocation,
		"timestamp":       a.Timestamp,
		"span_id":         a.SpanId,
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
	client    *logging.Service
	projectID string
}

func NewCloudLoggingClient(ctx context.Context, projectID string, credentialsJSON []byte) (*cloudLoggingClient, error) {
	var options []option.ClientOption
	if credentialsJSON != nil {
		options = append(options, option.WithCredentialsJSON(credentialsJSON))
	}
	service, err := logging.NewService(ctx, options...)
	if err != nil {
		return nil, err
	}

	return &cloudLoggingClient{
		client:    service,
		projectID: projectID,
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

func (c *cloudLoggingClient) ListLogEntries(ctx context.Context, filter string, limit int) ([]*Activity, error) {
	var entries []*Activity

	req := &logging.ListLogEntriesRequest{
		Filter:        filter,
		ResourceNames: []string{`projects/` + c.projectID},
	}
	if limit > 0 {
		req.OrderBy = "timestamp desc"
	}

	errLimitReached := errors.New("limit reached")
	if err := c.client.Entries.List(req).Pages(ctx, func(page *logging.ListLogEntriesResponse) error {
		for _, e := range page.Entries {
			if limit != 0 && len(entries) >= limit {
				return errLimitReached
			}
			entries = append(entries, &Activity{e})
		}
		return nil
	}); err != nil && err != errLimitReached {
		return nil, err
	}

	return entries, nil
}

func (c *cloudLoggingClient) GetLogBucket(ctx context.Context, name string) (*logging.LogBucket, error) {
	return c.client.Projects.Locations.Buckets.Get(name).Context(ctx).Do()
}
