package domain_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/stretchr/testify/assert"
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

func TestRequirementTrigger(t *testing.T) {
	t.Run("should return error if got error when parsing appeal to map", func(t *testing.T) {
		brokenAppeal := &domain.Appeal{
			Creator: map[string]interface{}{
				"foobar": &brokenType{},
			},
		}
		r := domain.RequirementTrigger{
			Expression: "$appeal.creator.foobar == true",
		}
		match, err := r.IsMatch(brokenAppeal)
		assert.ErrorContains(t, err, "parsing appeal to map:")
		assert.False(t, match)
	})

	t.Run("should return error if got error when evaluating expression", func(t *testing.T) {
		expr := "invalid expression"
		r := domain.RequirementTrigger{
			Expression: expr,
		}
		match, err := r.IsMatch(&domain.Appeal{})
		assert.ErrorContains(t, err, fmt.Sprintf("evaluating expression %q:", expr))
		assert.False(t, match)
	})

	t.Run("should return error if expression result is not boolean", func(t *testing.T) {
		expr := "$appeal.resource.details.foo"
		appeal := &domain.Appeal{
			Resource: &domain.Resource{
				Details: map[string]interface{}{
					"foo": "bar",
				},
			},
		}
		r := domain.RequirementTrigger{
			Expression: expr,
		}
		match, err := r.IsMatch(appeal)
		assert.ErrorContains(t, err, fmt.Sprintf("expression %q did not evaluate to a boolean, evaluated value: %q", expr, "bar"))
		assert.False(t, match)
	})

	t.Run("test trigger matching", func(t *testing.T) {
		testCases := []struct {
			name          string
			trigger       domain.RequirementTrigger
			appeal        *domain.Appeal
			expectedMatch bool
		}{
			{
				name: "using provider_type and resource_type",
				trigger: domain.RequirementTrigger{
					ProviderType: "my-provider",
					ResourceType: "my-resource",
				},
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Type:         "my-resource",
						ProviderType: "my-provider",
					},
				},
				expectedMatch: true,
			},
			{
				name: "check if resource_urn is matched expression",
				trigger: domain.RequirementTrigger{
					Expression: `$appeal.resource.urn == "urn:my-resource:123"`,
				},
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						URN: "urn:my-resource:123",
					},
				},
				expectedMatch: true,
			},
			{
				name: "check if resource_urn is matched expression (false condition)",
				trigger: domain.RequirementTrigger{
					Expression: `$appeal.resource.urn == "urn:my-resource:123"`,
				},
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						URN: "urn:my-resource:123456",
					},
				},
				expectedMatch: false,
			},
			{
				name: "expression returns boolean value",
				trigger: domain.RequirementTrigger{
					Expression: `$appeal.resource.details.foo`,
				},
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Details: map[string]interface{}{
							"foo": true,
						},
					},
				},
				expectedMatch: true,
			},
			{
				name: "using conditions",
				trigger: domain.RequirementTrigger{
					Conditions: []*domain.Condition{
						{
							Field: "$resource.type",
							Match: &domain.MatchCondition{
								Eq: "test-resource-type",
							},
						},
					},
				},
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Type: "test-resource-type",
					},
				},
				expectedMatch: true,
			},
			{
				name: "should return false on matched conditions and not matched expression",
				trigger: domain.RequirementTrigger{
					Conditions: []*domain.Condition{
						{
							Field: "$resource.type",
							Match: &domain.MatchCondition{
								Eq: "test-resource-type-incorrect",
							},
						},
					},
					Expression: `$appeal.resource.type == "test-resource-type"`,
				},
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Type: "test-resource-type",
					},
				},
				expectedMatch: false,
			},
			{
				name: "should return false on not matched conditions and matched expression",
				trigger: domain.RequirementTrigger{
					Conditions: []*domain.Condition{
						{
							Field: "$resource.type",
							Match: &domain.MatchCondition{
								Eq: "test-resource-type",
							},
						},
					},
					Expression: `$appeal.resource.type == "test-resource-type-incorrect"`,
				},
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Type: "test-resource-type",
					},
				},
				expectedMatch: false,
			},
			{
				name: "should return true on matched conditions and expression",
				trigger: domain.RequirementTrigger{
					Conditions: []*domain.Condition{
						{
							Field: "$resource.type",
							Match: &domain.MatchCondition{
								Eq: "test-resource-type",
							},
						},
					},
					Expression: `$appeal.resource.provider_type == "test-provider-type"`,
				},
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Type:         "test-resource-type",
						ProviderType: "test-provider-type",
					},
				},
				expectedMatch: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				match, err := tc.trigger.IsMatch(tc.appeal)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMatch, match)
			})
		}
	})
}

type brokenType struct{}

func (c *brokenType) MarshalJSON() ([]byte, error) {
	return []byte(""), fmt.Errorf("Error marshaling")
}
