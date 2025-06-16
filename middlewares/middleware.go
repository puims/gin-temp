package middlewares

import (
	"github.com/gin-gonic/gin"
)

func MiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 设置变量到Context的key中，可以通过Get()取
		c.Set("request", "中间件")

	}
}
