package utils

import (
	"errors"
	"gin-temp/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("happydays")

type Claims struct {
	ID       uint          `json:"uid"`
	Username string        `json:"username"`
	Roles    []models.Role `json:"roles"`
	jwt.RegisteredClaims
}

func GenerateToken(user *models.User) (gin.H, error) {
	expTime := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		ID:       user.ID,
		Username: user.Username,
		Roles:    user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			Issuer:    "gin-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		return gin.H{"error": "token_generation_failed"}, err
	}

	return gin.H{
		"token":   tokenStr,
		"expires": expTime.Format(time.RFC3339),
	}, nil
}

func GetToken(ctx *gin.Context) (token string, err error) {
	token = ctx.GetHeader("Authorization")

	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
		return
	} else {
		err = errors.New("failed to get token")
		return "", err
	}
}

func GetClaimsWithToken(ctx *gin.Context) (*Claims, error) {
	tokenStr, err := GetToken(ctx)
	if err != nil {
		ctx.JSON(401, gin.H{"error": "authorization header required"})
		ctx.Abort()
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
