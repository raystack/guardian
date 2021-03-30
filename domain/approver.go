package domain

import "time"

type Approver struct {
	ID         uint   `json:"id"`
	ApprovalID uint   `json:"approval_id"`
	AppealID   uint   `json:"appeal_id"`
	Email      string `json:"email"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
