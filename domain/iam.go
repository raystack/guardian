package domain

// ApproversResponse is the contract for the third-party service to fulfill the approver list request from guardian
type ApproversResponse struct {
	Emails []string `json:"emails"`
}

// IAMClient interface
type IAMClient interface {
	GetManagerEmails(user string) ([]string, error)
}

// IAMService interface
type IAMService interface {
	GetUserApproverEmails(user string) ([]string, error)
}
