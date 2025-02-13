package service

import (
	"crypto/sha256"
	"fmt"

	"github.com/go-faster/errors"

	"github.com/golang-jwt/jwt/v5"
)

const (
	salt       = "kqwemjksdnfhaksrmksvj283njwksdf"
	signingKey = "821nci1nc1234ubcz,mszd2jcv1wd23"
)

type tokenClaims struct {
	jwt.RegisteredClaims
	UserID int `json:"userID"`
}

func generatePasswordHash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}

func GenerateToken(username, password string) (string, error) {
	userID, err := stg.GetUser(username, generatePasswordHash(password))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		UserID: userID,
	})

	return token.SignedString([]byte(signingKey))
}

func ParseToken(accessToken string) (int, error) {

	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(signingKey), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok || claims == nil {
		return 0, errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserID, nil
}
