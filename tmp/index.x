package api

import (
	"net/http"
	"vercel-go/core/middleware"
	"vercel-go/core/types"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func Listen(w http.ResponseWriter, r *http.Request) {
	engine := gin.New()
	HandleRouter(engine)
	engine.ServeHTTP(w, r)
}

func HandleRouter(engine *gin.Engine) {
	baseRoute(engine)
	group := engine.Group("/api/v1")
	// router.SetupGinRoute(group)
	group.GET("/demo/index", func(context *gin.Context) {
		context.JSON(
			http.StatusOK,
			types.Success(
				"欢迎访问主页!",
				gin.H{
					"repository": "https://github.com/Fromsko/vercel-go",
					// "apiRoute":   appEngine.Routes(),
				},
			),
		)
	})
}

func baseRoute(engine *gin.Engine) {
	engine.Use(middleware.CORS_Default())
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
