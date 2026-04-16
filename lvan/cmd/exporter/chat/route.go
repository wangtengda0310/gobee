package chat

import "sync"

// RouteInfo 路由描述信息，用于自动生成帮助页面
type RouteInfo struct {
	Path        string   // 路由路径
	Method      string   // HTTP 方法（GET/POST/PUT/DELETE）
	Summary     string   // 一句话描述
	Description string   // 详细说明
	Params      []Param  // 请求参数定义
	Response    string   // 返回格式说明
}

// Param 请求参数定义
type Param struct {
	Name     string // 参数名
	Type     string // 类型（string, []string, ...）
	Required bool   // 是否必填
	Desc     string // 参数说明
}

var (
	chatRoutes []RouteInfo
	routeMu    sync.Mutex
)

// registerRoute 注册路由信息（并发安全）
func registerRoute(info RouteInfo) {
	routeMu.Lock()
	defer routeMu.Unlock()
	chatRoutes = append(chatRoutes, info)
}

// GetRoutes 获取所有已注册的路由信息
func GetRoutes() []RouteInfo {
	routeMu.Lock()
	defer routeMu.Unlock()
	result := make([]RouteInfo, len(chatRoutes))
	copy(result, chatRoutes)
	return result
}