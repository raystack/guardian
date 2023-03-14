package domain_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/goto/guardian/domain"
)

func TestStep_ResolveApprovers(t *testing.T) {
	tests := []struct {
		name    string
		appeal  *domain.Appeal
		step    domain.Step
		want    []string
		wantErr bool
	}{
		{
			name: "should resolve approvers",
			appeal: &domain.Appeal{
				Creator: map[string]interface{}{
					"userManager": "foo@bar.com",
				},
				Resource: &domain.Resource{
					Details: map[string]interface{}{
						"owner": "john.doe@example.com",
						"additionalOwners": []string{
							"moo@cow.fly",
							"foo@bar.app",
						},
					},
				},
			},
			step: domain.Step{
				Strategy: domain.ApprovalStepStrategyManual,
				Approvers: []string{
					"hello@world.id",
					"$appeal.creator.userManager",
					"$appeal.resource.details.owner",
					"$appeal.resource.details.additionalOwners",
				},
			},
			want: []string{
				"hello@world.id",
				"foo@bar.com",
				"john.doe@example.com",
				"moo@cow.fly",
				"foo@bar.app",
			},
			wantErr: false,
		},
		{
			name: "should return error if failed when evaluating expression",
			appeal: &domain.Appeal{
				Creator: map[string]interface{}{
					"userManager": "foo@bar.com",
				},
				Resource: &domain.Resource{
					Details: map[string]interface{}{
						"owner": "john.doe@example.com",
					},
				},
			},
			step: domain.Step{
				Strategy: domain.ApprovalStepStrategyManual,
				Approvers: []string{
					"hello@world.id",
					"$appeal.creator.userManager",
					"(*&)(#@*$&(#*&$)()))",
				},
			},
			wantErr: true,
		},
		{
			name: "should return error if value is not email",
			appeal: &domain.Appeal{
				Creator: map[string]interface{}{
					"userManager": "foo@bar.com",
				},
				Resource: &domain.Resource{
					Details: map[string]interface{}{
						"owner": "not-an-email",
					},
				},
			},
			step: domain.Step{
				Strategy: domain.ApprovalStepStrategyManual,
				Approvers: []string{
					"hello@world.id",
					"$appeal.creator.userManager",
					"$appeal.resource.details.owner",
				},
			},
			wantErr: true,
		},
		{
			name: "should return error if type is not supported",
			appeal: &domain.Appeal{
				Creator: map[string]interface{}{
					"userManager": "foo@bar.com",
				},
				Resource: &domain.Resource{
					Details: map[string]interface{}{
						"owner": 42,
					},
				},
			},
			step: domain.Step{
				Strategy: domain.ApprovalStepStrategyManual,
				Approvers: []string{
					"hello@world.id",
					"$appeal.creator.userManager",
					"$appeal.resource.details.owner",
				},
			},
			wantErr: true,
		},
		{
			name: "should return error if slice item type is not supported",
			appeal: &domain.Appeal{
				Creator: map[string]interface{}{
					"userManager": "foo@bar.com",
				},
				Resource: &domain.Resource{
					Details: map[string]interface{}{
						"owner": "john.doe@example.com",
						"additionalOwners": []interface{}{
							"moo@cow.fly",
							42,
						},
					},
				},
			},
			step: domain.Step{
				Strategy: domain.ApprovalStepStrategyManual,
				Approvers: []string{
					"hello@world.id",
					"$appeal.creator.userManager",
					"$appeal.resource.details.owner",
					"$appeal.resource.details.additionalOwners",
				},
			},
			wantErr: true,
		},
		{
			name: "should return error if appeal is unable to convert to map",
			appeal: &domain.Appeal{
				Creator: map[string]interface{}{
					"userManager": "foo@bar.com",
				},
				Resource: &domain.Resource{
					Details: map[string]interface{}{
						"owner":        "john.doe@example.com",
						"troubleMaker": &brokenType{},
					},
				},
			},
			step: domain.Step{
				Strategy: domain.ApprovalStepStrategyManual,
				Approvers: []string{
					"hello@world.id",
					"$appeal.creator.userManager",
					"$appeal.resource.details.owner",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.step.ResolveApprovers(tt.appeal)
			if (err != nil) != tt.wantErr {
				t.Errorf("Step.ResolveApprovers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Step.ResolveApprovers() = %v, want %v", got, tt.want)
			}
		})
	}
}

type brokenType struct{}

func (c *brokenType) MarshalJSON() ([]byte, error) {
	return []byte(""), fmt.Errorf("Error marshaling")
}
