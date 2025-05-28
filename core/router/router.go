package router

import (
	"net/http"
	"vercel-go/core/handler"
	"vercel-go/core/middleware"
	"vercel-go/core/types"

	"github.com/gin-gonic/gin"
)

func RegisterMiddleware(engine *gin.Engine) {
	engine.Use(middleware.CORS_Default())
}

func Setup(engine *gin.Engine) {
	// default
	engine.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "server is running!")
	})
	// notfound
	{
		engine.NoRoute(handler.NotFoundHandler)
		engine.NoMethod(handler.NotFoundHandler)
	}
	// middleware
	RegisterMiddleware(engine)
	// group router
	RegisterGroup(engine)
}

func RegisterGroup(engine *gin.Engine) {
	group := engine.Group("/api/v1")
	{
		group.GET("", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, types.Success("success", "/api/v1 router"))
		})
		group.POST("/trans", middleware.AuthMiddleware(), handler.TranslateHandler)
	}
}
