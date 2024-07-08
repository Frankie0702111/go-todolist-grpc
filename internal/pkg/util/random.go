package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/exp/constraints"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func RandomString(n int) string {
	sb := strings.Builder{}
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandomEmail() string {
	return fmt.Sprintf("%s@example.com", RandomString(6))
}

// Integer is a constraint that permits any integer type: int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64.
type Integer interface {
	constraints.Integer
}

// RandomInt returns a random integer between min and max (inclusive).
// It works with int, int32, and int64 types.
func RandomInt[T Integer](min, max T) T {
	// Ensure min is less than max
	if min > max {
		min, max = max, min
	}

	// Initialize the random number generator with a time-based seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random number in the range [0, max-min]
	return T(r.Int63n(int64(max-min+1))) + min
}

func RandomBool() bool {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(2) == 1
}
