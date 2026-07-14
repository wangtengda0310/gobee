package notification

// CheckResultHandler 检查结果输出通道接口
// 实现此接口可以将检查结果输出到不同的目标（控制台、飞书、文件等）
type CheckResultHandler interface {
	// Handle 处理检查结果事件
	// 返回 nil 表示成功，返回 error 表示处理失败
	Handle(event *CheckResultEvent) error

	// Name 返回通道名称（用于日志和调试）
	Name() string
}

// InterceptService 劫持服务接口（避免循环依赖）
type InterceptService interface {
	// AddMessage 添加被劫持的消息
	// 返回添加的消息，内部会通过事件推送到前端
	AddMessage(robotGUID, msgType, content string) InterceptedMessage
}

// InterceptedMessage 被劫持的消息
type InterceptedMessage struct {
	ID        string `json:"id"`        // 唯一ID
	RobotGUID string `json:"robotGuid"` // 机器人GUID
	MsgType   string `json:"msgType"`   // 消息类型：text / interactive
	Content   string `json:"content"`   // 消息内容
	Timestamp string `json:"timestamp"` // 发送时间（格式化后的字符串）
}

// EventsEmit 事件发送接口（避免循环依赖）
type EventsEmit func(name string, data any)
