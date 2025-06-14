package random

import (
	"crypto/rand"
	"math/big"
)

type Generator interface {
	FloatWithPrecision(min, max float64) (float64, error)
}

type CryptoGenerator struct{}

func NewCryptoGenerator() *CryptoGenerator {
	return &CryptoGenerator{}
}

func (g *CryptoGenerator) FloatWithPrecision(min, max float64) (float64, error) {
	minInt := int64(min * 100)
	maxInt := int64(max * 100)

	range_ := maxInt - minInt

	n := big.NewInt(range_ + 1)
	randomInt, err := rand.Int(rand.Reader, n)
	if err != nil {
		return 0, err
	}

	result := float64(randomInt.Int64()+minInt) / 100.0

	return result, nil
}
