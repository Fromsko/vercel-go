package interfaces

import "github.com/gin-gonic/gin"

type RegisterRouter interface {
	Setup(*gin.RouterGroup)
}
