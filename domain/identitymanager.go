package domain

// ApproversResponse is the contract for the third-party service to fulfill the approver list request from guardian
type ApproversResponse struct {
	Emails []string `json:"emails"`
}

// IdentityManagerClient interface
type IdentityManagerClient interface {
	GetUserApproverEmails(query map[string]string) ([]string, error)
}

// IdentityManagerService interface
type IdentityManagerService interface {
	GetUserApproverEmails(userEmail string) ([]string, error)
}
