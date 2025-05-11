package route

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"micro-mart/services/frontend_api/internal/api"
	"micro-mart/services/frontend_api/internal/config"
)

// RouteV1 is a struct implementing the Route interface
// Version 1 of the API
type RouteV1 struct {
	userHandler *api.UserHandler
	cfg         *config.Config
	// Add other handlers here
}

var _ Route = (*RouteV1)(nil)

// NewRouteV1Set creates a new fx.Option for the RouteV1 module
func NewRouteV1Set() fx.Option {
	return fx.Module("route-v1",
		fx.Provide(
			newRouteV1,
			api.NewUserHandler,
			// Add other handlers here
		),
	)
}

// newRouteV1 creates a new RouteV1 instance for the Route interface
func newRouteV1(
	cfg *config.Config,
	userHandler *api.UserHandler,
) Route {
	return &RouteV1{
		userHandler: userHandler,

		cfg: cfg,
	}
}

func (r *RouteV1) RegisterRoutes(g *gin.Engine) {

	v1 := g.Group("/api/v1")
	r.addUserRoutes(v1)
}

func (r *RouteV1) addUserRoutes(g *gin.RouterGroup) {
	auth := g.Group("/users")
	auth.POST("register", r.userHandler.Register)
}
