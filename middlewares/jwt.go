package middlewares

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("happydays")

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewToken(username string) (gin.H, error) {
	expTime := time.Now().Add(168 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			Issuer:    "gin-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		return gin.H{
			"error": "token_generation_failed",
		}, err
	}

	return gin.H{
		"token":   tokenStr,
		"expires": expTime.Format(time.RFC3339),
	}, nil
}

func Authorization(ctx *gin.Context) {
	tokenStr := ctx.GetHeader("Authorization")
	if tokenStr == "" {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "authorization header required"},
		)
		ctx.Abort()
		return
	}

	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "invalid token"},
		)
		ctx.Abort()
		return
	}

	ctx.Next()
}
