package util

import (
	"math/rand"
	"strings"
	"time"
)

func RandomOwner() string {
	return RandomString(5)
}

func RandomBalance() float64 {
	return RandomFloat(0, 10000)
}

func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "RUB", "UAH", "GBP", "BYN", "KZT"}
	max := int64(len(currencies) - 1)
	idx := RandomInt(0, max)
	return currencies[idx]
}

var randGen *rand.Rand

func init() {
	randSrc := rand.NewSource(time.Now().UnixNano())
	randGen = rand.New(randSrc)
}

func RandomInt(min, max int64) int64 {
	return min + randGen.Int63n(max-min+1)
}

func RandomFloat(min, max float64) float64 {
	return min + randGen.Float64()*(max-min+1)
}

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func RandomString(n int) string {
	strb := strings.Builder{}
	strb.Grow(n)

	const k = len(alphabet)

	for i := 0; i < n; i++ {
		_ = strb.WriteByte(
			alphabet[randGen.Intn(k)],
		)
	}

	return strb.String()
}
