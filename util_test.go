package disko

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindGaps(t *testing.T) {
	assert := assert.New(t)
	type data struct {
		expected uRanges
		ranges   uRanges
		min, max uint64
	}

	for _, d := range []data{
		{uRanges{{0, 100}}, uRanges{}, 0, 100},
		{uRanges{{0, 49}, {60, 100}}, uRanges{{50, 59}}, 0, 100},
		{uRanges{{51, 100}}, uRanges{{0, 50}}, 0, 100},
		{uRanges{{0, 10}}, uRanges{{11, 100}}, 0, 100},
		{uRanges{{0, 10}, {50, 59}, {91, 100}},
			uRanges{{11, 49}, {60, 90}}, 0, 100},
		{uRanges{}, uRanges{{0, 10}, {11, 100}}, 0, 100},
		{uRanges{}, uRanges{{0, 150}, {50, 100}}, 100, 100},
		{uRanges{{0, 9}, {41, 49}, {101, 110}},
			uRanges{{10, 40}, {50, 100}}, 0, 110},
		{uRanges{{10, 100}}, uRanges{{110, 200}}, 10, 100},
		{uRanges{{0, 9}, {51, 89}}, uRanges{{90, 100}, {10, 50}}, 0, 100},
		{uRanges{{2, 3}, {26, 52}, {62, 98}},
			uRanges{{0, 1}, {99, 100}, {53, 61}, {4, 25}}, 0, 100},
	} {
		assert.Equal(d.expected, findRangeGaps(d.ranges, d.min, d.max))
	}
}
