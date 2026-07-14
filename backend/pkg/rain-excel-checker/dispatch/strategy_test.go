package dispatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveNotifyMode_预发布分支(t *testing.T) {
	rules := DefaultRules()
	mode := ResolveNotifyMode("v0.0.8-pre-release", rules, DefaultMode())
	assert.True(t, mode.Group, "预发布分支应开启群消息")
	assert.False(t, mode.DM, "预发布分支应关闭私聊")
}

func TestResolveNotifyMode_普通分支(t *testing.T) {
	rules := DefaultRules()
	mode := ResolveNotifyMode("feature/hero-wiki", rules, DefaultMode())
	assert.False(t, mode.Group, "普通分支应关闭群消息")
	assert.True(t, mode.DM, "普通分支应开启私聊")
}

func TestResolveNotifyMode_空分支名(t *testing.T) {
	rules := DefaultRules()
	mode := ResolveNotifyMode("", rules, DefaultMode())
	assert.False(t, mode.Group)
	assert.True(t, mode.DM)
}

func TestResolveNotifyMode_main分支(t *testing.T) {
	rules := DefaultRules()
	mode := ResolveNotifyMode("main", rules, DefaultMode())
	assert.False(t, mode.Group)
	assert.True(t, mode.DM)
}

func TestResolveNotifyMode_自定义规则覆盖(t *testing.T) {
	rules := DefaultRules()
	rules["custom-branch"] = NotifyMode{Group: true, DM: true}

	mode := ResolveNotifyMode("custom-branch", rules, DefaultMode())
	assert.True(t, mode.Group)
	assert.True(t, mode.DM)

	// 未定义的分支仍走默认
	mode = ResolveNotifyMode("undefined", rules, DefaultMode())
	assert.False(t, mode.Group)
	assert.True(t, mode.DM)
}

func TestDefaultRules_包含预发布分支(t *testing.T) {
	rules := DefaultRules()
	_, exists := rules["v0.0.8-pre-release"]
	assert.True(t, exists, "DefaultRules 应包含 v0.0.8-pre-release")
}

func TestDefaultMode_值正确(t *testing.T) {
	mode := DefaultMode()
	assert.False(t, mode.Group, "默认模式应关闭群消息")
	assert.True(t, mode.DM, "默认模式应开启私聊")
}
