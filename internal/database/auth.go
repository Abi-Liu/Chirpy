package database

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func ComparePassword(password string, passwordToCompare string) error {
	err := bcrypt.CompareHashAndPassword([]byte(passwordToCompare), []byte(password))
	if err != nil {
		return err
	}

	return nil
}
