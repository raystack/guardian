package domain

import "time"

type Approver struct {
	ID         string `json:"id" yaml:"id"`
	ApprovalID string `json:"approval_id" yaml:"approval_id"`
	AppealID   string `json:"appeal_id" yaml:"appeal_id"`
	Email      string `json:"email" yaml:"email"`

	CreatedAt time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}
