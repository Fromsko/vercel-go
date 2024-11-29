package api

import (
	"context"
	v1 "vercel-go/handler/v1"

	"github.com/gin-gonic/gin"
)

type api struct {
	Ctx     context.Context
	Version string
}

func (a *api) Setup(group *gin.RouterGroup) {
	switch a.Version {
	case "v1":
		v1.NewAPI(group)
	}
}

func NewAPI(v string) *api {
	return &api{
		Ctx:     context.Background(),
		Version: v,
	}
}
