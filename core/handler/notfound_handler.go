package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NotFoundHandler(context *gin.Context) {
	context.JSON(
		http.StatusOK,
		gin.H{
			"code": 400,
			"msg":  "404 Not found",
			"err":  "The route is not defined.",
		},
	)
}
