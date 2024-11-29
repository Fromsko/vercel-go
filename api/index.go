package api

import (
	"net/http"
	"vercel-go/core/router"
)

func Listen(w http.ResponseWriter, r *http.Request) {
	engine := router.SetupGinRoute(
	// handler.NewAPI("v1"),
	)
	engine.ServeHTTP(w, r)
}
