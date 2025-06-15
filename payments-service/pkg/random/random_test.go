package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCryptoGenerator_Intn(t *testing.T) {
	gen := NewCryptoGenerator()
	n := 100

	for range 1000 {
		val := gen.Intn(n)
		assert.True(t, val >= 0 && val < n)
	}
}

func TestCryptoGenerator_Intn_Zero(t *testing.T) {
	gen := NewCryptoGenerator()
	val := gen.Intn(0)
	assert.Equal(t, 0, val)
}
