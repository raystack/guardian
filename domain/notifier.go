package domain

type NotificationMessages struct {
	ExpirationReminder   string `mapstructure:"expiration_reminder"`
	AppealApproved       string `mapstructure:"appeal_approved"`
	AppealRejected       string `mapstructure:"appeal_rejected"`
	AccessRevoked        string `mapstructure:"access_revoked"`
	ApproverNotification string `mapstructure:"approver_notification"`
	OthersAppealApproved string `mapstructure:"others_appeal_approved"`
	GrantOwnerChanged    string `mapstructure:"grant_owner_changed"`
}

const (
	NotificationTypeExpirationReminder     = "ExpirationReminder"
	NotificationTypeAppealApproved         = "AppealApproved"
	NotificationTypeOnBehalfAppealApproved = "OnBehalfAppealApproved"
	NotificationTypeAppealRejected         = "AppealRejected"
	NotificationTypeAccessRevoked          = "AccessRevoked"
	NotificationTypeApproverNotification   = "ApproverNotification"
	NotificationTypeGrantOwnerChanged      = "GrantOwnerChanged"
)

type NotificationMessage struct {
	Type      string
	Variables map[string]interface{}
}

type Notification struct {
	User    string
	Message NotificationMessage

	Labels map[string]string
}
