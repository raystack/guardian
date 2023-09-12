package evaluator_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goto/guardian/pkg/evaluator"
)

type ParamStruct struct {
	Bar string
}

func TestEvaluate(t *testing.T) {
	testCases := []struct {
		expression     string
		params         map[string]interface{}
		expectedResult interface{}
		expectedError  error
		aStruct        interface{}
	}{
		{
			expression:     "1 > 2",
			expectedResult: false,
		},
		{
			expression:     "5 == 2",
			expectedResult: false,
		},
		{
			expression:     `5 + 5 in [9,11,12]`,
			expectedResult: false,
		},
		{
			expression:     `"foo" == "bar"`,
			expectedResult: false,
		},
		{
			expression:    "$x",
			expectedError: evaluator.ErrExperssionParameterNotFound,
		},
		{
			expression: "$y",
			params: map[string]interface{}{
				"x": 1,
			},
			expectedError: evaluator.ErrExperssionParameterNotFound,
		},
		{
			expression: "$x > 1",
			params: map[string]interface{}{
				"x": 0,
			},
			expectedResult: false,
		},
		{
			expression: "$user.age > 10",
			params: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "john",
					"age":  10,
				},
			},
			expectedResult: false,
		},
		{
			expression: `$foo == "bar" && ($x == 1 && $y > $x)`,
			params: map[string]interface{}{
				"foo": "bar",
				"x":   1,
				"y":   2,
			},
			expectedResult: true,
		},
		{
			expression: "$foo.bar",
			params: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz",
				},
			},
			expectedResult: "baz",
		},
		{
			expression: "$user.name not in ['alpha', 'beta', 'gamma'] ? $user.name: 'one'",
			params: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "alpha",
				},
			},
			expectedResult: "one",
		},
		{
			expression: "!($user.email_id contains '@abc.com')",
			params: map[string]interface{}{
				"user": map[string]interface{}{
					"email_id": "user@example.com",
				},
			},
			expectedResult: true,
		},
		{
			// a complex test case to evaluate the expression where "the appeal role must either not contain viewer or (the resource type is gcloud_iam and the duration is 0h)
			expression: "!($appeal.role contains 'viewer') || $appeal.resource.provider_type == 'gcloud_iam' && $appeal.options.duration in ['0h','']",
			params: map[string]interface{}{
				"appeal": map[string]interface{}{
					"role": "Editor",
					"resource": map[string]interface{}{
						"provider_type": "gcloud_iam",
					},
					"options": map[string]interface{}{
						"duration": "0h",
					},
				},
			},
			expectedResult: true,
		},
		{
			// a complex test case to evaluate the expression where "the appeal role must either not contain viewer or (the resource type is gcloud_iam and the duration is 0h and the role must contain viewer)
			expression: "!($appeal.role contains 'viewer') || $appeal.resource.provider_type == 'gcloud_iam' && $appeal.options.duration in ['0h',''] && $appeal.role contains 'viewer'",
			params: map[string]interface{}{
				"appeal": map[string]interface{}{
					"role": "viewer",
					"resource": map[string]interface{}{
						"provider_type": "gcloud_iam",
					},
					"options": map[string]interface{}{
						"duration": "0h",
					},
				},
			},
			expectedResult: true,
		},
		{
			// a complex test case to evaluate the expression where "the appeal role must either not contain viewer or (the resource type is gcloud_iam and the duration is 0h and the role must contain viewer
			expression: "!($appeal.role contains 'viewer') || $appeal.resource.provider_type == 'gcloud_iam' && $appeal.options.duration in ['0h',''] && $appeal.role contains 'viewer'",
			params: map[string]interface{}{
				"appeal": map[string]interface{}{
					"role": "viewer",
					"resource": map[string]interface{}{
						"provider_type": "gcloud_iam",
					},
					"options": map[string]interface{}{
						"duration": "2160h",
					},
				},
			},
			expectedResult: false,
		},
		{
			// a complex test case to evaluate the expression where "the appeal role must either not contain admin_user or (the resource type is gcloud_iam and the duration is 0h and the role must contain admin_user)
			expression: "!($appeal.role contains 'admin_user') || $appeal.resource.provider_type == 'gcloud_iam' && $appeal.options.duration in ['0h',''] && $appeal.role contains 'admin_user'",
			params: map[string]interface{}{
				"appeal": map[string]interface{}{
					"role": "viewer",
					"resource": map[string]interface{}{
						"provider_type": "gcloud_iam",
					},
					"options": map[string]interface{}{
						"duration": "2160h",
					},
				},
			},
			expectedResult: true,
		},

		{
			expression: "len(Split($user.email_id, '@')[0])  > 2",
			params: map[string]interface{}{
				"user": map[string]interface{}{
					"email_id": "abc@example.com",
				},
			},
			expectedResult: true,
		},
		{
			expression: "$Bar",
			aStruct: ParamStruct{
				Bar: "baz",
			},
			expectedResult: "baz",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expression, func(t *testing.T) {
			e := evaluator.Expression(tc.expression)

			var actualResult interface{}
			var actualError error
			if tc.aStruct != nil {
				actualResult, actualError = e.EvaluateWithStruct(tc.aStruct)
			} else {
				actualResult, actualError = e.EvaluateWithVars(tc.params)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
			assert.ErrorIs(t, actualError, tc.expectedError)
		})
	}
}
