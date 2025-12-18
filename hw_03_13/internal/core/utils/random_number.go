package utils

import (
	"crypto/rand"
	"math/big"
)

// RandomAmount generates a random number in given interval, using
// cryptographically safe algorithm.
func RandomAmount(minVal, maxVal int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(maxVal-minVal+1)))
	return int(n.Int64()) + minVal
}
