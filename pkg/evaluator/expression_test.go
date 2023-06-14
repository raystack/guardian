package evaluator_test

import (
	"testing"

	"github.com/raystack/guardian/pkg/evaluator"
	"github.com/stretchr/testify/assert"
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
