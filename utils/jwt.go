package utils

import (
	"gin-temp/config"
	"gin-temp/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("happydays")

type Claims struct {
	ID       uint          `json:"uid"`
	Username string        `json:"username"`
	Roles    []models.Role `json:"roles"`
	jwt.RegisteredClaims
}

func GenerateToken(user *models.User) (string, error) {
	expTime := time.Now().Add(time.Duration(config.Viper.GetInt("app.expires")) * time.Hour)
	claims := &Claims{
		ID:       user.ID,
		Username: user.Username,
		Roles:    user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			Issuer:    "gin-app",
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtKey)

	return token, err
}

func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}
