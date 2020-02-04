package disko

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindGaps0(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{0, 100}},
		findRangeGaps([]uRange{}, 0, 100))
}

func TestFindGaps1(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{0, 49}, {60, 100}},
		findRangeGaps([]uRange{{50, 59}}, 0, 100))
}

func TestFindGaps2(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{51, 100}},
		findRangeGaps([]uRange{{0, 50}}, 0, 100))
}

func TestFindGaps3(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{0, 10}},
		findRangeGaps([]uRange{{11, 100}}, 0, 100))
}

func TestFindGaps4(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{0, 10}, {50, 59}, {91, 100}},
		findRangeGaps([]uRange{{11, 49}, {60, 90}}, 0, 100))
}

func TestFindGaps5(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{},
		findRangeGaps([]uRange{{0, 10}, {11, 100}}, 0, 100))
}

func TestFindGaps6(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{},
		findRangeGaps([]uRange{{0, 150}, {50, 100}}, 100, 100))
}

func TestFindGaps7(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{0, 9}, {41, 49}, {101, 110}},
		findRangeGaps([]uRange{{10, 40}, {50, 100}}, 0, 110))
}

func TestFindGaps8(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{10, 100}},
		findRangeGaps([]uRange{{110, 200}}, 10, 100))
}
