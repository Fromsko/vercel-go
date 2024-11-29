package router

import (
	"net/http"
	"vercel-go/core/interfaces"
	"vercel-go/core/middleware"

	"github.com/gin-gonic/gin"
)

var appEngine *gin.Engine

func init() {
	gin.SetMode(
		gin.ReleaseMode,
	)
	appEngine = gin.Default()
}

func register(engine *gin.Engine, group *gin.RouterGroup, routes ...interfaces.RegisterRouter) *gin.Engine {
	withDefaultMiddleware(engine)
	withBaseRoute(engine)

	for _, route := range routes {
		route.Setup(group)
	}

	return engine
}

func withDefaultMiddleware(engine *gin.Engine) {
	engine.Use(
		middleware.CORS_Default(),
	)
}

func withBaseRoute(engine *gin.Engine) {
	engine.NoRoute(func(context *gin.Context) {
		context.JSON(
			http.StatusOK,
			gin.H{
				"code": 400,
				"msg":  "404 Not found",
				"err":  "The route is not defined.",
			},
		)
	})
}

func SetupGinRoute(routes ...interfaces.RegisterRouter) *gin.Engine {
	appEngine.GET("/demo/index", func(context *gin.Context) {
		context.JSON(
			http.StatusOK,
			gin.H{
				"code": 200,
				"msg":  "欢迎访问主页!",
				"data": gin.H{
					"repository": "https://github.com/Fromsko/vercel-go",
					// "apiRoute":   appEngine.Routes(),
				},
			},
		)
	})

	appEngine.NoRoute(func(context *gin.Context) {
		context.JSON(
			http.StatusOK,
			gin.H{
				"code": 400,
				"msg":  "404 Not found",
				"err":  "The route is not defined.",
			},
		)
	})

	// baseGroup := appEngine.Group("")

	// register(
	// 	appEngine,
	// 	baseGroup,
	// 	routes...,
	// )

	return appEngine
}
