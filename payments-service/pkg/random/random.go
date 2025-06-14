package random

import (
	"crypto/rand"
	"math/big"
)

type Generator interface {
	Intn(n int) int
}

type CryptoGenerator struct{}

func NewCryptoGenerator() *CryptoGenerator {
	return &CryptoGenerator{}
}

func (g *CryptoGenerator) Intn(n int) int {
	if n <= 0 {
		return 0
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}

	return int(nBig.Int64())
}
