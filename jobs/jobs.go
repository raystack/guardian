package jobs

import "github.com/mitchellh/mapstructure"

type Type string

const (
	TypeFetchResources             Type = "fetch_resources"
	TypeExpiringGrantNotification  Type = "expiring_grant_notification"
	TypeRevokeExpiredGrants        Type = "revoke_expired_grants"
	TypeRevokeGrantsByUserCriteria Type = "revoke_grants_by_user_criteria"
	TypeGrantDormancyCheck         Type = "grant_dormancy_check"
)

type Job struct {
	Type   Type
	Config Config `mapstructure:"config"`
}

// Config is a map of job-specific configuration
type Config map[string]interface{}

func (c Config) Decode(v interface{}) error {
	return mapstructure.Decode(c, v)
}
