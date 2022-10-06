package utils_test

import (
	"testing"

	"github.com/odpf/guardian/utils"
	"github.com/stretchr/testify/assert"
)

func TestContainsOrdered(t *testing.T) {
	testCases := []struct {
		slice             []string
		lookingFor        []string
		expectedResult    bool
		expectedHeadIndex int
	}{
		{
			[]string{"a", "b", "c", "d"},
			[]string{"a", "b", "c", "d"},
			true, 0,
		},
		{
			[]string{"a", "b", "c", "d"},
			[]string{"a", "b"},
			true, 0,
		},
		{
			[]string{"a", "b", "c", "d"},
			[]string{"b", "c", "d"},
			true, 1,
		},
		{
			// return true for empty `lookingFor` elements
			[]string{"a", "b", "c", "d"},
			[]string{},
			true, 0,
		},
		{
			[]string{},
			[]string{},
			true, 0,
		},
		{
			[]string{"a", "b", "c", "d"},
			[]string{"a", "c", "d"},
			false, 0,
		},
		{
			[]string{"a", "b", "c", "d"},
			[]string{"c", "c", "d"},
			false, 2,
		},
		{
			[]string{"a", "b", "c", "d"},
			[]string{"b", "d"},
			false, 1,
		},
		{
			[]string{"a", "b", "c", "d"},
			[]string{"a", "b", "x"},
			false, 0,
		},
		{
			[]string{"a", "b", "c", "d"},
			[]string{"a", "b", "c", "d", "x", "y"},
			false, 0,
		},
		{
			[]string{},
			[]string{"a"},
			false, 0,
		},
	}

	for _, tc := range testCases {
		result, headIndex := utils.ContainsOrdered(tc.slice, tc.lookingFor)
		assert.Equal(t, tc.expectedResult, result)
		assert.Equal(t, tc.expectedHeadIndex, headIndex)
	}
}
