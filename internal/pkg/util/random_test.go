package util_test

import (
	"go-todolist-grpc/internal/pkg/util"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	t.Run("CorrectLength", func(t *testing.T) {
		lengths := []int{0, 1, 5, 10, 20}
		for _, length := range lengths {
			result := util.RandomString(length)
			assert.Len(t, result, length)
		}
	})

	t.Run("OnlyLowercaseLetters", func(t *testing.T) {
		result := util.RandomString(100)
		assert.Regexp(t, "^[a-z]+$", result)
	})

	t.Run("RandomnessCheck", func(t *testing.T) {
		result1 := util.RandomString(100)
		result2 := util.RandomString(100)
		assert.NotEqual(t, result1, result2)
	})

	t.Run("ZeroLength", func(t *testing.T) {
		result := util.RandomString(0)
		assert.Empty(t, result)
	})

	t.Run("LargeLength", func(t *testing.T) {
		length := 1000000
		result := util.RandomString(length)
		assert.Equal(t, length, len(result))
	})
}

func TestRandomEmail(t *testing.T) {
	t.Run("ValidEmailFormat", func(t *testing.T) {
		email := util.RandomEmail()
		assert.Regexp(t, `^[a-z]{6}@example\.com$`, email)
	})

	t.Run("UniqueEmails", func(t *testing.T) {
		email1 := util.RandomEmail()
		email2 := util.RandomEmail()
		assert.NotEqual(t, email1, email2)
	})

	t.Run("ConsistentDomain", func(t *testing.T) {
		email := util.RandomEmail()
		parts := strings.Split(email, "@")
		assert.Equal(t, "example.com", parts[1])
	})

	t.Run("LocalPartLength", func(t *testing.T) {
		email := util.RandomEmail()
		localPart := strings.Split(email, "@")[0]
		assert.Equal(t, 6, len(localPart))
	})
}
