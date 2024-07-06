package util_test

import (
	"go-todolist-grpc/internal/pkg/util"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		jwtTTL := 15
		jwtSecretKey := "mysecretkey"
		userID := int(123)

		token, err := util.GenerateToken(jwtTTL, jwtSecretKey, userID)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims := &util.CustomClaims{}
		_, _, err = new(jwt.Parser).ParseUnverified(token, claims)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.WithinDuration(t, time.Now().Add(time.Minute*time.Duration(jwtTTL)), time.Unix(claims.ExpiresAt, 0), time.Second)
	})
}

func TestParseToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		jwtTTL := 15
		jwtSecretKey := "mysecretkey"
		userID := int(123)

		token, err := util.GenerateToken(jwtTTL, jwtSecretKey, userID)
		assert.NoError(t, err)

		claims, err := util.ParseToken(jwtSecretKey, token)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.WithinDuration(t, time.Now().Add(time.Minute*time.Duration(jwtTTL)), time.Unix(claims.ExpiresAt, 0), time.Second)
	})

	t.Run("Failure_InvalidToken", func(t *testing.T) {
		jwtSecretKey := "mysecretkey"
		invalidToken := "invalidtoken"

		claims, err := util.ParseToken(jwtSecretKey, invalidToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Failure_InvalidSigningMethod", func(t *testing.T) {
		jwtSecretKey := "mysecretkey"
		invalidToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

		claims, err := util.ParseToken(jwtSecretKey, invalidToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
