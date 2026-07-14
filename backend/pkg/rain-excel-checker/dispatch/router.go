package dispatch

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification/handlers"
)

// NotifyRouter 封装 Dispatcher 的构建逻辑
// 根据通知模式决定注册哪些 handler
type NotifyRouter struct {
	mode      NotifyMode
	robotGUID string
	dmHandler notification.CheckResultHandler // 可为 nil（未配置私聊凭证或 -noDM 时）
}

// NewNotifyRouter 根据通知模式创建 router
// mode: 通知模式（由 ResolveNotifyMode 决定）
// robotGUID: 飞书机器人 webhook ID
// dmHandler: 私聊 handler（未配置时传 nil，可被装饰器包装）
func NewNotifyRouter(mode NotifyMode, robotGUID string, dmHandler notification.CheckResultHandler) *NotifyRouter {
	return &NotifyRouter{
		mode:      mode,
		robotGUID: robotGUID,
		dmHandler: dmHandler,
	}
}

// BuildDispatcher 构建 Dispatcher 并注册对应 handler
// authors: 用于群消息 @ 提及的用户名列表
func (r *NotifyRouter) BuildDispatcher(authors []string) *notification.CheckResultDispatcher {
	dispatcher := notification.NewDispatcher()

	// 控制台输出始终注册
	dispatcher.Register(handlers.NewConsoleHandler())

	// 按模式注册群消息 handler
	if r.mode.Group {
		opts := handlerAuthorsOptions(authors)
		dispatcher.Register(handlers.NewFeishuCardHandler(r.robotGUID, opts...))
	}

	// 按模式注册私聊 handler（可能被装饰器包装，包含 debug 监控）
	if r.mode.DM && r.dmHandler != nil {
		dispatcher.Register(r.dmHandler)
	}

	return dispatcher
}

// Mode 返回当前通知模式
func (r *NotifyRouter) Mode() NotifyMode {
	return r.mode
}

// handlerAuthorsOptions 根据作者列表构建飞书卡片选项
func handlerAuthorsOptions(authors []string) []handlers.FeishuCardOption {
	switch len(authors) {
	case 0:
		return nil
	case 1:
		return []handlers.FeishuCardOption{handlers.WithAtUser(authors[0])}
	default:
		return []handlers.FeishuCardOption{handlers.WithAtUsers(authors)}
	}
}
