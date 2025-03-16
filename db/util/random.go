package util

import (
	"math/rand"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func RandomOwner() string {
	return RandomString(5)
}

func RandomBalance() pgtype.Numeric {
	n, err := Float64ToNumeric(RandomFloat(0, 10000))
	if err != nil {
		panic(err)
	}
	return n
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
		strb.WriteByte(
			alphabet[randGen.Intn(k)],
		)
	}

	return strb.String()
}
