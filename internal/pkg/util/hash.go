package util

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(cost int, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)

	return string(hash), err
}

func CheckPasswordHash(password string, hashPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))

	return err == nil
}
