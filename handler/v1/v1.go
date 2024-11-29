package v1

import "github.com/gin-gonic/gin"

var (
	Group *gin.RouterGroup
	Path  = "/api/v1"
)

func NewAPI(gp *gin.RouterGroup) {
	Group = gp.Group(Path)
}
