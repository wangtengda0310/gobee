package prototest

// EventEmitter 定义事件发射能力，与 Wails v3 application.EventManager 解耦。
// record_worker.go / replay_worker.go 通过此接口推送事件，避免直接依赖
// github.com/wailsapp/wails/v3/pkg/application，使 proto-test 包在 CLI/无 CGO 场景下仍可编译。
type EventEmitter interface {
	Emit(name string, data ...any) bool
}
