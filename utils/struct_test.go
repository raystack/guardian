package utils_test

import (
	"testing"

	"github.com/raystack/guardian/utils"
	"github.com/stretchr/testify/assert"
)

type testStructToMap struct {
	Key        string `json:"key"`
	OmittedKey string `json:"omitted_key,omitempty"`
}

func TestStructToMap(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected map[string]interface{}
	}{
		{
			name: "should return map with all fields",
			input: testStructToMap{
				Key:        "value",
				OmittedKey: "value",
			},
			expected: map[string]interface{}{
				"key":         "value",
				"omitted_key": "value",
			},
		},
		{
			name: "should return map with omitted fields",
			input: testStructToMap{
				Key: "value",
			},
			expected: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name:     "should return empty map when input is nil",
			input:    nil,
			expected: map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := utils.StructToMap(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			assert.Equal(t, tc.expected, result)
		})
	}
}
