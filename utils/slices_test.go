package utils_test

import (
	"testing"

	"github.com/odpf/guardian/utils"
	"github.com/stretchr/testify/assert"
)

func TestSubsliceExists(t *testing.T) {
	testCases := []struct {
		slice             []string
		subslice          []string
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
			[]string{"a", "b", "c"},
			[]string{"b"},
			true, 1,
		},
		{
			[]string{"a", "b", "c", "b"},
			[]string{"c"},
			true, 2,
		},
		{
			[]string{"a", "b", "c", "d"},
			[]string{"b", "c", "d"},
			true, 1,
		},
		{
			// return true for empty `subslice` elements
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
			false, 0,
		},
		{
			[]string{"a", "b", "c", "d"},
			[]string{"b", "d"},
			false, 0,
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
		result, headIndex := utils.SubsliceExists(tc.slice, tc.subslice)
		assert.Equal(t, tc.expectedResult, result)
		assert.Equal(t, tc.expectedHeadIndex, headIndex)
	}
}
