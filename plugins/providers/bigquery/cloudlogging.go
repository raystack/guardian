package bigquery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/logging"
	logadmin "cloud.google.com/go/logging/logadmin"
	"github.com/odpf/guardian/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/cloud/audit"
)

var (
	ErrInvalidPayload = errors.New("payload is not of type *audit.AuditLog")
	ErrPayloadIsNil   = errors.New("couldn't get payload from log entry")
)

type auditLog audit.AuditLog

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

type logEntry logging.Entry

func (e logEntry) GetAuditLog() (*auditLog, error) {
	l, ok := e.Payload.(*audit.AuditLog)
	if !ok {
		return nil, ErrInvalidPayload
	}
	return (*auditLog)(l), nil
}

func (e logEntry) ToProviderAcivity(p domain.Provider) (*domain.ProviderActivity, error) {
	a := &domain.ProviderActivity{
		ProviderID: p.ID,
		Timestamp:  e.Timestamp,
	}

	al, err := e.GetAuditLog()
	if err != nil {
		return nil, err
	}
	if al == nil {
		return nil, ErrPayloadIsNil
	}

	a.Type = al.MethodName
	if al.AuthenticationInfo != nil {
		a.AccountType = al.GetAccountType()
		a.AccountID = al.AuthenticationInfo.PrincipalEmail
	}

	if payloadJson, err := json.Marshal(al); err != nil {
		return nil, fmt.Errorf("marshalling payload: %w", err)
	} else if err := json.Unmarshal(payloadJson, &a.Metadata); err != nil {
		return nil, fmt.Errorf("unmarshalling payload to metadata: %w", err)
	}

	for _, ai := range al.AuthorizationInfo {
		a.Authorizations = append(a.Authorizations, ai.Permission)
	}

	a.Resource = al.GetResource(p)

	return a, nil
}

type cloudLoggingClient struct {
	client *logadmin.Client
}

func NewCloudLoggingClient(ctx context.Context, projectID string) (*cloudLoggingClient, error) {
	client, err := logadmin.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &cloudLoggingClient{
		client: client,
	}, nil
}

func (c *cloudLoggingClient) ListLogEntries(ctx context.Context, provider domain.Provider) ([]*domain.ProviderActivity, error) {
	lastHour := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	filter := fmt.Sprintf(`
protoPayload.serviceName="bigquery.googleapis.com" AND
timestamp > "%s"
	`, lastHour) // TODO: make this configurable

	var activities []*domain.ProviderActivity
	it := c.client.Entries(ctx, logadmin.Filter(filter))
	for len(activities) < 5 {
		e, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, fmt.Errorf("iterating over cloud logging entries: %w", err)
		}

		en := (*logEntry)(e)
		a, err := en.ToProviderAcivity(provider)
		if err != nil {
			return nil, fmt.Errorf("converting cloud logging entry to provider activity: %w", err)
		}
		activities = append(activities, a)
	}

	return activities, nil
}

func (c *cloudLoggingClient) Close() error {
	return c.client.Close()
}
