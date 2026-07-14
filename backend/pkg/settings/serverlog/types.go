// internal/serverlog/types.go
package serverlog

// LogEntry 服务端日志条目，通过 serverLog 事件发送到前端
type LogEntry struct {
	Level     string `json:"level"`     // DEBUG | INFO | WARN | ERROR
	Message   string `json:"message"`   // 日志内容
	Timestamp string `json:"timestamp"` // 时间戳，格式 15:04:05.000
	IsManual  bool   `json:"isManual"`  // 是否为手动标记的日志
}

// ToMap 转换为 map[string]any 用于 Wails Event.Emit
// Wails v3 的 Event.Emit 无法正确序列化自定义 Go 结构体，
// 但 map[string]any 可以正常工作
func (e LogEntry) ToMap() map[string]any {
	return map[string]any{
		"level":     e.Level,
		"message":   e.Message,
		"timestamp": e.Timestamp,
		"isManual":  e.IsManual,
	}
}
