package evaluator_test

import (
	"testing"

	"github.com/odpf/guardian/evaluator"
	"github.com/stretchr/testify/assert"
)

func TestEvaluate(t *testing.T) {
	testCases := []struct {
		expression     string
		params         map[string]interface{}
		expectedResult interface{}
		expectedError  error
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
	}

	for _, tc := range testCases {
		t.Run(tc.expression, func(t *testing.T) {
			e := evaluator.Expression(tc.expression)

			actualResult, actualError := e.EvaluateWithVars(tc.params)

			assert.Equal(t, tc.expectedResult, actualResult)
			assert.ErrorIs(t, actualError, tc.expectedError)
		})
	}
}
