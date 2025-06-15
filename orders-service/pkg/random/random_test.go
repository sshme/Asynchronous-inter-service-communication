package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCryptoGenerator_FloatWithPrecision(t *testing.T) {
	gen := NewCryptoGenerator()
	min := 10.0
	max := 20.0

	val, err := gen.FloatWithPrecision(min, max)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, val, min)
	assert.LessOrEqual(t, val, max)
}
