package domain

type NotificationMessages struct {
	ExpirationReminder   string `mapstructure:"expiration_reminder" default:"[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"Your access {{.account_id}} to {{.resource_name}} with role {{.role}} will expire at {{.expiration_date}}. Extend the access if it's still needed\"}}]"`
	AppealApproved       string `mapstructure:"appeal_approved" default:"[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"Your appeal to {{.resource_name}} with role {{.role}} has been approved\"}}]"`
	AppealRejected       string `mapstructure:"appeal_rejected" default:"[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"Your appeal to {{.resource_name}} with role {{.role}} has been rejected\"}}]"`
	AccessRevoked        string `mapstructure:"access_revoked" default:"[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"Your access to {{.resource_name}}} with role {{.role}} has been revoked\"}}]"`
	ApproverNotification string `mapstructure:"approver_notification" default:"[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"You have an appeal created by {{.requestor}} requesting access to {{.resource_name}} with role {{.role}}.\\n Appeal ID: {{.appeal_id}}\"}}]"`
	OthersAppealApproved string `mapstructure:"others_appeal_approved" default:"[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"Your appeal to {{.resource_name}} with role {{.role}} created by {{.requestor}} has been approved\"}}]"`
}

const (
	NotificationTypeExpirationReminder     = "ExpirationReminder"
	NotificationTypeAppealApproved         = "AppealApproved"
	NotificationTypeOnBehalfAppealApproved = "OnBehalfAppealApproved"
	NotificationTypeAppealRejected         = "AppealRejected"
	NotificationTypeAccessRevoked          = "AccessRevoked"
	NotificationTypeApproverNotification   = "ApproverNotification"
)

type NotificationMessage struct {
	Type      string
	Variables map[string]interface{}
}

type Notification struct {
	User    string
	Message NotificationMessage
}
