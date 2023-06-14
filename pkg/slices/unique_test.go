package slices_test

import (
	"testing"

	"github.com/raystack/guardian/pkg/slices"
	"github.com/stretchr/testify/assert"
)

func TestUniqueStringSlice(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "one",
			input:    []string{"a"},
			expected: []string{"a"},
		},
		{
			name:     "double b",
			input:    []string{"a", "b", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "double c",
			input:    []string{"c", "b", "c"},
			expected: []string{"c", "b"},
		},
		{
			name:     "complex",
			input:    []string{"b", "b", "c", "a", "b", "c", "a", "b", "c"},
			expected: []string{"b", "c", "a"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := slices.UniqueStringSlice(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
