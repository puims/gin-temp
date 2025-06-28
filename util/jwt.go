package util

import (
	"context"
	"gin-temp/model"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var (
	JwtKey        []byte
	JwtKeyRefresh []byte
)

type Claims struct {
	ID       uint   `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func initJWTKeys() {
	JwtKey = []byte(Viper.GetString("jwt.access_key"))
	JwtKeyRefresh = []byte(Viper.GetString("jwt.refresh_key"))

	if len(JwtKey) < 32 || len(JwtKeyRefresh) < 32 {
		log.Fatal("JWT keys must be at least 32 bytes long")
	}
}

func GenerateToken(user *model.User, expires int, jwtKey []byte) (string, error) {
	expTime := time.Now().Add(time.Duration(expires) * time.Hour)
	claims := &Claims{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			Issuer:    "gin-app",
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtKey)

	return token, err
}

func ParseToken(token string, jwtKey []byte) (*Claims, error) {
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

func AddToBlackList(token string, expiresTime time.Time, rds *redis.Client) error {
	expiration := time.Until(expiresTime)
	return rds.Set(
		context.Background(),
		"blacklist:"+token,
		"1",
		expiration).Err()
}

func InBlackList(token string, rds *redis.Client) (bool, error) {
	_, err := rds.Get(context.Background(), "blacklist:"+token).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
