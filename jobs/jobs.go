package jobs

import "github.com/mitchellh/mapstructure"

type Type string

const (
	TypeFetchResources             Type = "fetch_resources"
	TypeExpiringGrantNotification  Type = "expiring_grant_notification"
	TypeRevokeExpiredGrants        Type = "revoke_expired_grants"
	TypeRevokeGrantsByUserCriteria Type = "revoke_grants_by_user_criteria"
	TypeGrantDormancyCheck         Type = "grant_dormancy_check"

	// Deprecated: use RevokeExpiredGrants instead
	TypeRevokeExpiredAccess Type = "revoke_expired_access"
	// Deprecated: use ExpiringGrantNotification instead
	TypeExpiringAccessNotification Type = "expiring_access_notification"
)

type Job struct {
	Type Type
	// Enabled is set as true for backward compatibility. If the job needs to be disabled, it must be present in the config with this value as false.
	Enabled  bool   `mapstructure:"enabled" default:"true"`
	Interval string `mapstructure:"interval"`
	Config   Config `mapstructure:"config"`
}

// Config is a map of job-specific configuration
type Config map[string]interface{}

func (c Config) Decode(v interface{}) error {
	return mapstructure.Decode(c, v)
}
