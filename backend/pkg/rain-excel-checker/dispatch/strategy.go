// Package dispatch 提供通知分发策略和路由功能
package dispatch

// NotifyMode 分支策略决定的通知通道开关
type NotifyMode struct {
	Group bool // 是否发群卡片消息
	DM    bool // 是否发飞书私聊消息
}

// ResolveNotifyMode 根据分支名和预定义规则返回通知模式
// branch: 当前 git 分支名
// rules: 分支名 → NotifyMode 的映射规则
// defaultMode: 未匹配任何规则时使用的默认模式
func ResolveNotifyMode(branch string, rules map[string]NotifyMode, defaultMode NotifyMode) NotifyMode {
	if mode, ok := rules[branch]; ok {
		return mode
	}
	return defaultMode
}

// DefaultRules 返回预定义的分支通知规则
func DefaultRules() map[string]NotifyMode {
	return map[string]NotifyMode{
		"v0.0.8-pre-release": {Group: true, DM: false},
	}
}

// DefaultMode 返回默认通知模式（非预定义分支使用）
func DefaultMode() NotifyMode {
	return NotifyMode{Group: false, DM: true}
}
