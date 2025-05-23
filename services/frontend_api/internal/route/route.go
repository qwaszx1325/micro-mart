package route

import "github.com/gin-gonic/gin"

// Route is an interface for registering routes
type Route interface {
	RegisterRoutes(g *gin.Engine)
}
