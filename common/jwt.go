package common

import (
	"dning.com/pro02/model"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var jwtKey = []byte("a_secret_crect")

type Claims struct {
	SID string
	jwt.RegisteredClaims
}

type Claims2 struct {
	TID string
	jwt.RegisteredClaims
}

func ReleaseToken(student model.Student) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		SID: student.SID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "DN",
			Subject:   "user token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ReleaseToken2(teacher model.Teacher) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims2{
		TID: teacher.TID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "DN",
			Subject:   "user token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenStr string) (*jwt.Token, *Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	return token, claims, err
}

func ParseToken2(tokenStr string) (*jwt.Token, *Claims2, error) {
	claims2 := &Claims2{}
	token, err := jwt.ParseWithClaims(tokenStr, claims2, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	return token, claims2, err
}
