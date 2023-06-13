package indexmap

import (
	"reflect"
	"testing"
)

func mapsEqual(x, y map[int]int) bool {
	return reflect.DeepEqual(x, y)
}
func TestIndexMap_Set(t *testing.T) {
	testCases := []struct {
		name                string
		indexTransformation func(IndexMap) IndexMap
		input               IndexMap
		expected            IndexMap
	}{
		{
			name: "empty indexmap as input and output",
			indexTransformation: func(input IndexMap) IndexMap {
				input.Set(0, 0)
				return input
			},
			input:    IndexMap{},
			expected: IndexMap{},
		},
		{
			name: "map single index to a different index",
			indexTransformation: func(input IndexMap) IndexMap {
				input.Set(1, 2)
				return input
			},
			input:    IndexMap{},
			expected: IndexMap{1: 2},
		},
		{
			name: "map multiple indices to different indices",
			indexTransformation: func(input IndexMap) IndexMap {
				input.Set(1, 2)
				input.Set(3, 4)
				return input
			},
			input:    IndexMap{},
			expected: IndexMap{1: 2, 3: 4},
		},
		{
			name: "map some indices and leave others",
			indexTransformation: func(input IndexMap) IndexMap {
				input.Set(1, 2)
				return input
			},
			input:    IndexMap{3: 3},
			expected: IndexMap{1: 2, 3: 3},
		},
		{
			name: "map an index to itself",
			indexTransformation: func(input IndexMap) IndexMap {
				input.Set(1, 1)
				return input
			},
			input:    IndexMap{},
			expected: IndexMap{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.indexTransformation(tc.input)
			if !mapsEqual(actual, tc.expected) {
				t.Errorf("expected %v but got %v", tc.expected, actual)
			}
		})
	}
}

func TestIndexMap_Get(t *testing.T) {
	testCases := []struct {
		name         string
		baseIndexMap IndexMap
		input        int
		expected     int
	}{
		{
			name:         "empty indexmap",
			baseIndexMap: IndexMap{},
			input:        1,
			expected:     1,
		},
		{
			name:         "indexmap with mapped indices, querying mapped index",
			baseIndexMap: IndexMap{1: 2, 3: 4},
			input:        1,
			expected:     2,
		},
		{
			name:         "indexmap with mapped indices, querying non-mapped index",
			baseIndexMap: IndexMap{1: 2, 3: 4},
			input:        5,
			expected:     5,
		},
		{
			name:         "non-empty indexmap without queried index",
			baseIndexMap: IndexMap{3: 3},
			input:        1,
			expected:     1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.baseIndexMap.Get(tc.input)
			if actual != tc.expected {
				t.Errorf("expected %v but got %v", tc.expected, actual)
			}
		})
	}
}
