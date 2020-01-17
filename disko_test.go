package disko_test

import (
	"testing"

	"github.com/anuvu/disko"
	"github.com/stretchr/testify/assert"
)

func TestFreeSpace(t *testing.T) {
	assert := assert.New(t)

	values := [][]uint64{
		{10, 10, 0},
		{0, 200, 200},
		{100, 200, 100},
	}

	for _, v := range values {
		f := disko.FreeSpace{v[0], v[1]}
		assert.Equal(f.Size(), v[2])
	}
}
