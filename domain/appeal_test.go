package domain_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/goto/guardian/domain"
	"github.com/stretchr/testify/assert"
)

func TestAppeal_GetNextPendingApproval(t *testing.T) {
	tests := []struct {
		name   string
		appeal domain.Appeal
		want   *domain.Approval
	}{
		{
			name: "should return nil if no approvals",
			appeal: domain.Appeal{
				Approvals: []*domain.Approval{},
			},
			want: nil,
		},
		{
			name: "should return pending approval if exists",
			appeal: domain.Appeal{
				Approvals: []*domain.Approval{
					{
						ID:        "1",
						Status:    domain.ApprovalStatusApproved,
						Approvers: []string{"user1"},
					},
					{
						ID:        "2",
						Status:    domain.ApprovalStatusPending,
						Approvers: []string{"user1"},
					},
				},
			},
			want: &domain.Approval{
				ID:        "2",
				Status:    domain.ApprovalStatusPending,
				Approvers: []string{"user1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appeal.GetNextPendingApproval(); !assert.Equal(t, got, tt.want) {
				t.Errorf("Appeal.GetNextPendingApproval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppeal_Init(t *testing.T) {
	a := domain.Appeal{}
	p := &domain.Policy{
		ID:      "policy-1",
		Version: 1,
	}
	a.Init(p)

	assert.Equal(t, a.Status, domain.AppealStatusPending)
	assert.Equal(t, a.PolicyID, p.ID)
	assert.Equal(t, a.PolicyVersion, p.Version)
}

func TestAppeal_Cancel(t *testing.T) {
	a := domain.Appeal{}
	a.Cancel()

	assert.Equal(t, a.Status, domain.AppealStatusCanceled)
}

func TestAppeal_Reject(t *testing.T) {
	a := domain.Appeal{}
	a.Reject()

	assert.Equal(t, a.Status, domain.AppealStatusRejected)
}

func TestAppeal_Approve(t *testing.T) {
	tests := []struct {
		name    string
		appeal  *domain.Appeal
		checks  func(t *testing.T, a *domain.Appeal)
		wantErr bool
	}{
		{
			name:   "should change status to approved",
			appeal: &domain.Appeal{},
			checks: func(t *testing.T, a *domain.Appeal) {
				t.Helper()
				assert.Equal(t, a.Status, domain.AppealStatusApproved)
			},
			wantErr: false,
		},
		{
			name: "should return error if duration is not valid",
			appeal: &domain.Appeal{
				Options: &domain.AppealOptions{
					Duration: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "should be able to approve with permanent access",
			appeal: &domain.Appeal{
				Options: &domain.AppealOptions{
					Duration: "0",
				},
			},
			checks: func(t *testing.T, a *domain.Appeal) {
				t.Helper()
				assert.Equal(t, a.Status, domain.AppealStatusApproved)
			},
			wantErr: false,
		},
		{
			name: "should be able to approve with temporary access",
			appeal: &domain.Appeal{
				Options: &domain.AppealOptions{
					Duration: "1h",
				},
			},
			checks: func(t *testing.T, a *domain.Appeal) {
				t.Helper()
				assert.Equal(t, a.Status, domain.AppealStatusApproved)
				oneHourLater := time.Now().Add(1 * time.Hour)
				assert.GreaterOrEqual(t, oneHourLater, *a.Options.ExpirationDate)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.appeal.Approve(); (err != nil) != tt.wantErr {
				t.Errorf("Appeal.Approve() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.checks != nil {
				tt.checks(t, tt.appeal)
			}
		})
	}
}

func TestAppeal_SetDefaults(t *testing.T) {
	tests := []struct {
		name   string
		appeal *domain.Appeal
		checks func(t *testing.T, a *domain.Appeal)
	}{
		{
			name:   "should set default values if account type is not set",
			appeal: &domain.Appeal{},
			checks: func(t *testing.T, a *domain.Appeal) {
				t.Helper()
				assert.Equal(t, a.AccountType, domain.DefaultAppealAccountType)
			},
		},
		{
			name: "should set default values if account type is set",
			appeal: &domain.Appeal{
				AccountType: "test",
			},
			checks: func(t *testing.T, a *domain.Appeal) {
				t.Helper()
				assert.Equal(t, a.AccountType, "test")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.appeal.SetDefaults()
			tt.checks(t, tt.appeal)
		})
	}
}

func TestAppeal_GetApproval(t *testing.T) {
	tests := []struct {
		name   string
		appeal *domain.Appeal
		id     string
		want   *domain.Approval
	}{
		{
			name: "should return approval with given id",
			appeal: &domain.Appeal{
				Approvals: []*domain.Approval{
					{
						ID: "approval-1",
					},
				},
			},
			id: "approval-1",
			want: &domain.Approval{
				ID: "approval-1",
			},
		},
		{
			name: "should return nil if approval with given id does not exist",
			appeal: &domain.Appeal{
				Approvals: []*domain.Approval{
					{
						ID: "approval-1",
					},
				},
			},
			id:   "non-existing",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appeal.GetApproval(tt.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Appeal.GetApproval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppeal_ToGrant(t *testing.T) {
	tests := []struct {
		name    string
		appeal  domain.Appeal
		want    *domain.Grant
		wantErr bool
	}{
		{
			name: "should return permanent grant",
			appeal: domain.Appeal{
				ID:          "appeal-1",
				AccountID:   "account-1",
				AccountType: "test",
				ResourceID:  "resource-1",
				Role:        "role-1",
				Permissions: []string{"permission-1"},
				CreatedBy:   "user-1",
			},
			want: &domain.Grant{
				Status:      domain.GrantStatusActive,
				AccountID:   "account-1",
				AccountType: "test",
				ResourceID:  "resource-1",
				Role:        "role-1",
				Permissions: []string{"permission-1"},
				AppealID:    "appeal-1",
				CreatedBy:   "user-1",
				IsPermanent: true,
			},
			wantErr: false,
		},
		{
			name: "should return permanent grant if duration is zero",
			appeal: domain.Appeal{
				ID:          "appeal-1",
				AccountID:   "account-1",
				AccountType: "test",
				ResourceID:  "resource-1",
				Role:        "role-1",
				Permissions: []string{"permission-1"},
				CreatedBy:   "user-1",
				Options: &domain.AppealOptions{
					Duration: "0",
				},
			},
			want: &domain.Grant{
				Status:      domain.GrantStatusActive,
				AccountID:   "account-1",
				AccountType: "test",
				ResourceID:  "resource-1",
				Role:        "role-1",
				Permissions: []string{"permission-1"},
				AppealID:    "appeal-1",
				CreatedBy:   "user-1",
				IsPermanent: true,
			},
			wantErr: false,
		},
		{
			name: "should return temporary grant if duration is not zero",
			appeal: domain.Appeal{
				ID:          "appeal-1",
				AccountID:   "account-1",
				AccountType: "test",
				ResourceID:  "resource-1",
				Role:        "role-1",
				Permissions: []string{"permission-1"},
				CreatedBy:   "user-1",
				Options: &domain.AppealOptions{
					Duration: "1h",
				},
			},
			want: &domain.Grant{
				Status:      domain.GrantStatusActive,
				AccountID:   "account-1",
				AccountType: "test",
				ResourceID:  "resource-1",
				Role:        "role-1",
				Permissions: []string{"permission-1"},
				AppealID:    "appeal-1",
				CreatedBy:   "user-1",
				IsPermanent: false,
			},
			wantErr: false,
		},
		{
			name: "should return error if invalid duration",
			appeal: domain.Appeal{
				ID:          "appeal-1",
				AccountID:   "account-1",
				AccountType: "test",
				ResourceID:  "resource-1",
				Role:        "role-1",
				Permissions: []string{"permission-1"},
				CreatedBy:   "user-1",
				Options: &domain.AppealOptions{
					Duration: "invalid",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.appeal.ToGrant()
			if (err != nil) != tt.wantErr {
				t.Errorf("Appeal.ToGrant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == false && tt.want.IsPermanent == false {
				tt.want.ExpirationDate = got.ExpirationDate
			}
			if !assert.Equal(t, got, tt.want) {
				t.Errorf("Appeal.ToGrant() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppeal_AdvanceApproval(t *testing.T) {
	tests := []struct {
		name          string
		appeal        *domain.Appeal
		wantErr       bool
		wantApprovals []*domain.Approval
	}{
		{
			name: "should resolve multiple automatic approval steps",
			appeal: &domain.Appeal{
				PolicyID:      "test-id",
				PolicyVersion: 1,
				Resource: &domain.Resource{
					Name: "grafana",
					Details: map[string]interface{}{
						"owner": "test-owner",
					},
				},
				Policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:      "step-1",
							ApproveIf: `$appeal.resource.details.owner == "test-owner"`,
						},
						{
							Name:      "step-2",
							ApproveIf: `$appeal.resource.details.owner == "test-owner"`,
						},
						{
							Name:      "step-3",
							ApproveIf: `$appeal.resource.details.owner == "test-owner"`,
						},
					},
				},
				Approvals: []*domain.Approval{
					{
						Status: "pending",
						Index:  0,
					},
					{
						Status: "blocked",
						Index:  1,
					},
					{
						Status: "blocked",
						Index:  2,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should autofill rejection reason on auto-reject",
			appeal: &domain.Appeal{
				PolicyID:      "test-id",
				PolicyVersion: 1,
				Resource: &domain.Resource{
					Name: "grafana",
					Details: map[string]interface{}{
						"owner": "test-owner",
					},
				},
				Policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:            "step-1",
							Strategy:        "auto",
							RejectionReason: "test rejection reason",
							ApproveIf:       `false`, // hard reject for testing purpose
						},
					},
				},
				Approvals: []*domain.Approval{
					{
						Status: domain.ApprovalStatusPending,
						Index:  0,
					},
				},
			},
			wantErr: false,
			wantApprovals: []*domain.Approval{
				{
					Status: domain.ApprovalStatusRejected,
					Index:  0,
					Reason: "test rejection reason",
				},
			},
		},
		{
			name: "should do nothing if approvals is already rejected",
			appeal: &domain.Appeal{
				PolicyID:      "test-id",
				PolicyVersion: 1,
				Resource: &domain.Resource{
					Name: "grafana",
					Details: map[string]interface{}{
						"owner": "test-owner",
					},
				},
				Policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:            "step-1",
							Strategy:        "auto",
							RejectionReason: "test rejection reason",
							ApproveIf:       `false`, // hard reject for testing purpose
						},
					},
				},
				Approvals: []*domain.Approval{
					{
						Status: domain.AppealStatusRejected,
						Index:  0,
					},
				},
			},
			wantErr: false,
			wantApprovals: []*domain.Approval{
				{
					Status: domain.ApprovalStatusRejected,
					Index:  0,
				},
			},
		},
		{
			name: "should return error if invalid expression",
			appeal: &domain.Appeal{
				PolicyID:      "test-id",
				PolicyVersion: 1,
				Resource: &domain.Resource{
					Name: "grafana",
					Details: map[string]interface{}{
						"owner": "test-owner",
					},
				},
				Policy: &domain.Policy{
					ID:      "test-id",
					Version: 1,
					Steps: []*domain.Step{
						{
							Name:      "step-1",
							Strategy:  "auto",
							ApproveIf: `)*(&_#)($U#_)(`, // invalid expression
						},
					},
				},
				Approvals: []*domain.Approval{
					{
						Status: domain.AppealStatusPending,
						Index:  0,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.appeal.AdvanceApproval(tt.appeal.Policy); (err != nil) != tt.wantErr {
				t.Errorf("Appeal.AdvanceApproval() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantApprovals != nil {
				assert.Equal(t, tt.wantApprovals, tt.appeal.Approvals)
			}
		})
	}
}

func TestAppeal_AdvanceApproval_UpdateApprovalStatuses(t *testing.T) {
	resourceFlagStep := &domain.Step{
		Name: "resourceFlagStep",
		When: "$appeal.resource.details.flag == true",
		Approvers: []string{
			"user@email.com",
		},
	}
	humanApprovalStep := &domain.Step{
		Name: "humanApprovalStep",
		Approvers: []string{
			"human@email.com",
		},
	}

	testCases := []struct {
		name                     string
		appeal                   *domain.Appeal
		steps                    []*domain.Step
		existingApprovalStatuses []string
		expectedApprovalStatuses []string
		expectedErrorStr         string
	}{
		{
			name: "initial process, When on the first step",
			appeal: &domain.Appeal{
				Resource: &domain.Resource{
					Details: map[string]interface{}{
						"flag": false,
					},
				},
			},
			steps: []*domain.Step{
				resourceFlagStep,
				humanApprovalStep,
			},
			existingApprovalStatuses: []string{
				domain.ApprovalStatusPending,
				domain.ApprovalStatusBlocked,
			},
			expectedApprovalStatuses: []string{
				domain.ApprovalStatusSkipped,
				domain.ApprovalStatusPending,
			},
		},
		{
			name: "When expression fulfilled",
			appeal: &domain.Appeal{
				Resource: &domain.Resource{
					Details: map[string]interface{}{
						"flag": true,
					},
				},
			},
			steps: []*domain.Step{
				humanApprovalStep,
				resourceFlagStep,
				humanApprovalStep,
			},
			existingApprovalStatuses: []string{
				domain.ApprovalStatusApproved,
				domain.ApprovalStatusPending,
				domain.ApprovalStatusBlocked,
			},
			expectedApprovalStatuses: []string{
				domain.ApprovalStatusApproved,
				domain.ApprovalStatusPending,
				domain.ApprovalStatusBlocked,
			},
		},
		{
			name: "should access nested fields properly in expression",
			appeal: &domain.Appeal{
				Resource: &domain.Resource{},
			},
			steps: []*domain.Step{
				{
					Strategy:  "manual",
					When:      `$appeal.details != nil && $appeal.details.foo != nil && $appeal.details.bar != nil && ($appeal.details.foo.foo contains "foo" || $appeal.details.foo.bar contains "bar")`,
					Approvers: []string{"approver1@email.com"},
				},
				{
					Strategy:  "manual",
					Approvers: []string{"approver2@email.com"},
				},
			},
			existingApprovalStatuses: []string{
				domain.ApprovalStatusPending,
				domain.ApprovalStatusBlocked,
			},
			expectedApprovalStatuses: []string{
				domain.ApprovalStatusSkipped,
				domain.ApprovalStatusPending,
			},
		},
		{
			name: "should return error if failed when evaluating expression",
			appeal: &domain.Appeal{
				Resource: &domain.Resource{},
			},
			steps: []*domain.Step{
				{
					Strategy:  "manual",
					When:      `$appeal.details != nil && $appeal.details.foo != nil && $appeal.details.bar != nil && $appeal.details.foo.foo contains "foo" || $appeal.details.foo.bar contains "bar"`,
					Approvers: []string{"approver1@email.com"},
				},
				{
					Strategy:  "manual",
					Approvers: []string{"approver2@email.com"},
				},
			},
			existingApprovalStatuses: []string{
				domain.ApprovalStatusPending,
				domain.ApprovalStatusPending,
			},
			expectedErrorStr: "evaluating expression ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appeal := *tc.appeal
			for i, s := range tc.existingApprovalStatuses {
				appeal.Approvals = append(appeal.Approvals, &domain.Approval{
					Status: s,
					Index:  i,
				})
			}
			appeal.Policy = &domain.Policy{
				Steps: tc.steps,
			}
			actualError := appeal.AdvanceApproval(appeal.Policy)
			if tc.expectedErrorStr == "" {
				assert.Nil(t, actualError)
				for i, a := range appeal.Approvals {
					assert.Equal(t, a.Status, tc.expectedApprovalStatuses[i])
				}
			} else {
				assert.Contains(t, actualError.Error(), tc.expectedErrorStr)
			}
		})
	}
}

func TestAppeal_ApplyPolicy(t *testing.T) {
	tests := []struct {
		name          string
		appeal        *domain.Appeal
		policy        *domain.Policy
		wantApprovals []*domain.Approval
		wantErr       bool
	}{
		{
			name:   "should return no approvals if steps are empty",
			appeal: &domain.Appeal{},
			policy: &domain.Policy{
				Steps: []*domain.Step{},
			},
			wantApprovals: []*domain.Approval{},
			wantErr:       false,
		},
		{
			name:   "should return correct approvals",
			appeal: &domain.Appeal{},
			policy: &domain.Policy{
				Steps: []*domain.Step{
					{
						Strategy:  domain.ApprovalStepStrategyAuto,
						ApproveIf: `1 == 1`,
					},
					{
						Strategy:  domain.ApprovalStepStrategyManual,
						Approvers: []string{"john.doe@example.com"},
					},
				},
			},
			wantApprovals: []*domain.Approval{
				{
					Index:  0,
					Status: domain.ApprovalStatusPending,
				},
				{
					Index:     1,
					Status:    domain.ApprovalStatusBlocked,
					Approvers: []string{"john.doe@example.com"},
				},
			},
			wantErr: false,
		},
		{
			name:   "should return error if failed to resolve approvers",
			appeal: &domain.Appeal{},
			policy: &domain.Policy{
				Steps: []*domain.Step{
					{
						Strategy:  domain.ApprovalStepStrategyManual,
						Approvers: []string{")*(@#&$_(*)#$&)(*"},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.appeal.ApplyPolicy(tt.policy); (err != nil) != tt.wantErr {
				t.Errorf("Appeal.ApplyPolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.appeal.Approvals, tt.wantApprovals)
		})
	}
}
