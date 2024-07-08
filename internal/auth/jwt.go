package auth

import (
	"log"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(secret string, id, expiresAt int) (string, error) {
	dayInSeconds := 24 * 60 * 60
	log.Print("expires start: ", expiresAt)
	if expiresAt == 0 || expiresAt > dayInSeconds {
		expiresAt = dayInSeconds
	}

	log.Print("expires final: ", expiresAt)

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(expiresAt) * time.Second)),
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		Subject:   strconv.Itoa(id),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", err
	}

	return ss, nil
}

func ParseToken(tokenString, secret string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		log.Fatal("Unkown claims")
	}

	return subject, nil
}
