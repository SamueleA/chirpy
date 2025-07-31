package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytesString := []byte(password)

	hashedPassword, err := bcrypt.GenerateFromPassword(bytesString, bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err
}