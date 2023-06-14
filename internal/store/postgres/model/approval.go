package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/guardian/domain"
	"gorm.io/gorm"
)

// Approval database model
type Approval struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name          string    `gorm:"index"`
	Index         int
	AppealID      string
	Status        string
	Actor         *string
	Reason        string
	PolicyID      string
	PolicyVersion uint

	Approvers []Approver
	Appeal    *Appeal

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// FromDomain transforms *domain.Approval values into the model
func (m *Approval) FromDomain(a *domain.Approval) error {
	var approvers []Approver
	if a.Approvers != nil {
		for _, approver := range a.Approvers {
			m := new(Approver)
			if err := m.FromDomain(&domain.Approver{Email: approver}); err != nil {
				return err
			}
			approvers = append(approvers, *m)
		}
	}

	if a.Appeal != nil {
		appealModel := new(Appeal)
		if err := appealModel.FromDomain(a.Appeal); err != nil {
			return err
		}
		m.Appeal = appealModel
	}

	var id uuid.UUID
	if a.ID != "" {
		uuid, err := uuid.Parse(a.ID)
		if err != nil {
			return fmt.Errorf("parsing uuid: %w", err)
		}
		id = uuid
	}
	m.ID = id
	m.Name = a.Name
	m.Index = a.Index
	m.AppealID = a.AppealID
	m.Status = a.Status
	m.Actor = a.Actor
	m.Reason = a.Reason
	m.PolicyID = a.PolicyID
	m.PolicyVersion = a.PolicyVersion
	m.Approvers = approvers
	m.CreatedAt = a.CreatedAt
	m.UpdatedAt = a.UpdatedAt

	return nil
}

// ToDomain transforms model into *domain.Approval
func (m *Approval) ToDomain() (*domain.Approval, error) {
	var approvers []string
	if m.Approvers != nil {
		for _, a := range m.Approvers {
			approver := a.ToDomain()
			approvers = append(approvers, approver.Email)
		}
	}

	var appeal *domain.Appeal
	if m.Appeal != nil {
		a, err := m.Appeal.ToDomain()
		if err != nil {
			return nil, err
		}
		appeal = a
	}

	return &domain.Approval{
		ID:            m.ID.String(),
		Name:          m.Name,
		Index:         m.Index,
		AppealID:      m.AppealID,
		Status:        m.Status,
		Actor:         m.Actor,
		Reason:        m.Reason,
		PolicyID:      m.PolicyID,
		PolicyVersion: m.PolicyVersion,
		Approvers:     approvers,
		Appeal:        appeal,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}
