package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 检查 x-scr 请求头是否合法
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 x-scr 请求头
		scrHeader := c.GetHeader("x-scr")

		// 检查是否合法
		if scrHeader != "fromsko-666" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid x-scr header",
			})
			c.Abort()
			return
		}

		// 合法，继续处理
		c.Next()
	}
}
