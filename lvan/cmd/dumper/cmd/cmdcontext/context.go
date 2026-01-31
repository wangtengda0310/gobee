package cmdcontext

import (
	"context"

	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

type contextKey string

const (
	ManagerKey contextKey = "datasource-manager"
)

// SetManager 设置数据源管理器到 context
func SetManager(ctx context.Context, mgr service.Manager) context.Context {
	return context.WithValue(ctx, ManagerKey, mgr)
}

// GetManager 从 context 获取数据源管理器
func GetManager(ctx context.Context) service.Manager {
	if mgr, ok := ctx.Value(ManagerKey).(service.Manager); ok {
		return mgr
	}
	return nil
}
