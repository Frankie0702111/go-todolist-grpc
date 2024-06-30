package util_test

import (
	"go-todolist-grpc/internal/pkg/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		password := "mysecretpassword"
		cost := 4

		hashedPassword, err := util.HashPassword(cost, password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
	})

	t.Run("Failure_InvalidPassword", func(t *testing.T) {
		password := "mypasswordmypasswordmypasswordmypasswordmypasswordmypasswordmypasswordmypassword"
		cost := 4

		hashedPassword, err := util.HashPassword(cost, password)
		assert.Error(t, err, "bcrypt: password length exceeds 72 bytes")
		assert.Empty(t, hashedPassword)
	})
}

func TestCheckPasswordHash(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		password := "mysecretpassword"
		cost := 4

		hashedPassword, err := util.HashPassword(cost, password)
		assert.NoError(t, err)

		isValid := util.CheckPasswordHash(password, hashedPassword)
		assert.True(t, isValid)
	})

	t.Run("Failure_InvalidPassword", func(t *testing.T) {
		password := "mysecretpassword"
		wrongPassword := "wrongpassword"
		cost := 4

		hashedPassword, err := util.HashPassword(cost, password)
		assert.NoError(t, err)

		isValid := util.CheckPasswordHash(wrongPassword, hashedPassword)
		assert.False(t, isValid)
	})
}
