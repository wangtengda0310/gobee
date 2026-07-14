package functiontest

// RobotLogEmitter 定义机器人日志事件发射能力，与 Wails v3 application.EventManager 解耦。
// services.go 的 RunRobotTest 通过此接口推送 "robotLog" 事件，避免直接依赖
// github.com/wailsapp/wails/v3/pkg/application，使本包在 CLI/无 CGO 场景下仍可编译。
//
// Wails 的 *application.EventManager 实现了 Emit(name string, data ...any) bool，
// 因此 GUI 入口可直接传入 app.Event；CLI/MCP 入口传 nil（emitRobotLog 内有 nil 守卫）。
type RobotLogEmitter interface {
	Emit(name string, data ...any) bool
}
