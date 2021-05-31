package domain

// IAMClient interface
type IAMClient interface {
	GetManagerEmails(user string) ([]string, error)
}

// IAMService interface
type IAMService interface {
	GetUserApproverEmails(user string) ([]string, error)
}
