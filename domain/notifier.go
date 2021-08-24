package domain

import "time"

type Notifier interface {
	Notify([]Notification) error
}

type NotificationMessages struct {
	ExpirationReminder   string `mapstructure:"expiration_reminder" default:"Your access to {{resource_name}} with role {{role}} will expire at {{expiration_date}}. Extend the access if it's still needed"`
	AppealApproved       string `mapstructure:"appeal_approved" default:"Your appeal to {{resource_name}} with role {{role}} has been approved"`
	AppealRejected       string `mapstructure:"appeal_rejected" default:"Your appeal to {{resource_name}} with role {{role}} has been rejected"`
	AccessRevoked        string `mapstructure:"access_revoked" default:"Your access to {{resource_name}}} with role {{role}} has been revoked"`
	ApproverNotification string `mapstructure:"approver_notification" default:"You have an appeal created by {{requestor}} requesting access to {{resource_name}} with role {{role}}. Appeal ID: {{appeal_id}}"`
}

const (
	NotificationTypeExpirationReminder   = "ExpirationReminder"
	NotificationTypeAppealApproved       = "AppealApproved"
	NotificationTypeAppealRejected       = "AppealRejected"
	NotificationTypeAccessRevoked        = "AccessRevoked"
	NotificationTypeApproverNotification = "ApproverNotification"
)

type NotificationVariables struct {
	ResourceName   string
	Role           string
	ExpirationDate time.Time
	Requestor      string
	AppealID       uint
}

type NotificationMessage struct {
	Type      string
	Variables NotificationVariables
}

type Notification struct {
	User    string
	Message NotificationMessage
}
