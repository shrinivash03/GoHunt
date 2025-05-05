package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthClaims struct {
	Id            string `json:"id"`
	User          string `json:"user"`
	Admin         bool   `json:"role"`
	jwt.RegisteredClaims `json:"claims"`
}

func CreateNewAuthToken(id string, email string, isAdmin bool) (string, error) {
	claims := AuthClaims{
		Id: id,
		User: email,
		Admin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer: "gohunt.com",
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secretKey, exists := os.LookupEnv("SECRET_KEY")
	if !exists {
		panic("SECRET_KEY cannot be found in .env")
	}

	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", errors.New("error signing the token")
	}
	return signedToken, nil
}