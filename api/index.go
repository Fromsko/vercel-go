package api

import (
	"net/http"
	"vercel-go/core/router"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	engine := gin.New()
	router.Setup(engine)
	engine.ServeHTTP(w, r)
}
