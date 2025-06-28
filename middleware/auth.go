package middleware

import (
	"errors"
	"gin-temp/model"
	"gin-temp/util"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func JwtAuthorization() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")

		isblack, err := util.InBlackList(token, util.Redis)
		if err != nil {
			model.ErrorAbortResponse(ctx, 500, err)
			return
		}
		if isblack {
			model.ErrorAbortResponse(ctx, 401, errors.New("user has logout"))
			return
		}

		if len(token) < 7 || token[:7] != "Bearer " {
			model.ErrorAbortResponse(ctx, 401, errors.New("failed to get token from header"))
			return
		}
		token = token[7:]

		claims, err := util.ParseToken(token, util.JwtKey)
		if err != nil {
			model.ErrorAbortResponse(ctx, 401, err)
			return
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}

func CasbinAuthorization(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取请求的URI
		obj := ctx.Request.URL.RequestURI()
		// 获取请求方法
		act := ctx.Request.Method
		// 获取用户的角色
		claims, exists := ctx.Get("claims")
		if !exists {
			model.ErrorAbortResponse(ctx, 401, errors.New("unauthorized"))
			return
		}
		sub := claims.(*util.Claims).Role

		ok, err := enforcer.Enforce(sub, obj, act)
		if err != nil {
			model.ErrorAbortResponse(ctx, 500, err)
			return
		}
		if !ok {
			model.ErrorAbortResponse(ctx, 403, errors.New("forbidden"))
			return
		}

		ctx.Next()
	}
}
