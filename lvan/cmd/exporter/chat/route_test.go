package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterRoute(t *testing.T) {
	chatRoutes = nil

	registerRoute(RouteInfo{
		Path:    "/chat/api/test",
		Method:  "GET",
		Summary: "测试路由",
	})

	assert.Equal(t, 1, len(chatRoutes))
	assert.Equal(t, "/chat/api/test", chatRoutes[0].Path)
	assert.Equal(t, "GET", chatRoutes[0].Method)
	assert.Equal(t, "测试路由", chatRoutes[0].Summary)
}

func TestRegisterRouteMultiple(t *testing.T) {
	chatRoutes = nil

	registerRoute(RouteInfo{Path: "/a", Method: "GET", Summary: "A"})
	registerRoute(RouteInfo{Path: "/b", Method: "POST", Summary: "B"})
	registerRoute(RouteInfo{Path: "/c", Method: "DELETE", Summary: "C"})

	assert.Equal(t, 3, len(chatRoutes))
}

func TestGetRoutes(t *testing.T) {
	chatRoutes = nil

	registerRoute(RouteInfo{Path: "/x", Method: "GET", Summary: "X"})
	routes := GetRoutes()

	assert.Equal(t, 1, len(routes))
	assert.Equal(t, "/x", routes[0].Path)
}