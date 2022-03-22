package domain

type NotificationMessages struct {
	ExpirationReminder   string `mapstructure:"expiration_reminder" default:"Your access to {{.resource_name}} with role {{.role}} will expire at {{.expiration_date}}. Extend the access if it's still needed, click this link to view {{.console_url}}/dataaccess/resources."`
	AppealApproved       string `mapstructure:"appeal_approved" default:"Your appeal to {{.resource_name}} with role {{.role}} has been approved, click this link to view {{.console_url}}/dataaccess/requests."`
	AppealRejected       string `mapstructure:"appeal_rejected" default:"Your appeal to {{.resource_name}} with role {{.role}} has been rejected, click this link to view {{.console_url}}/dataaccess/requests."`
	AccessRevoked        string `mapstructure:"access_revoked" default:"Your access to {{.resource_name}}} with role {{.role}} has been revoked, click this link to view {{.console_url}}/dataaccess/requests."`
	ApproverNotification string `mapstructure:"approver_notification" default:"You have an appeal created by {{.requestor}} requesting access to {{.resource_name}} with role {{.role}}. Appeal ID: {{.appeal_id}}, click this link to view {{.console_url}}/dataaccess/manage-requests."`
}

const (
	NotificationTypeExpirationReminder   = "ExpirationReminder"
	NotificationTypeAppealApproved       = "AppealApproved"
	NotificationTypeAppealRejected       = "AppealRejected"
	NotificationTypeAccessRevoked        = "AccessRevoked"
	NotificationTypeApproverNotification = "ApproverNotification"
)

type NotificationMessage struct {
	Type      string
	Variables map[string]interface{}
}

type Notification struct {
	User    string
	Message NotificationMessage
}
