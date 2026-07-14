package dispatch

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification"
	"github.com/stretchr/testify/assert"
)

// noopHandler 测试用 mock handler
type noopHandler struct{}

func (noopHandler) Handle(_ *notification.CheckResultEvent) error { return nil }
func (noopHandler) Name() string                                  { return "noop" }

func TestBuildDispatcher_群消息模式(t *testing.T) {
	router := NewNotifyRouter(NotifyMode{Group: true, DM: false}, "test-robot", nil)
	d := router.BuildDispatcher([]string{"张三"})
	// Console + FeishuCard = 2
	assert.Equal(t, 2, d.HandlerCount(), "群消息模式应注册 Console + Card")
}

func TestBuildDispatcher_私聊模式有DM(t *testing.T) {
	router := NewNotifyRouter(NotifyMode{Group: false, DM: true}, "test-robot", noopHandler{})
	d := router.BuildDispatcher([]string{"张三"})
	// Console + DM = 2
	assert.Equal(t, 2, d.HandlerCount(), "私聊模式(有DM)应注册 Console + DM")
}

func TestBuildDispatcher_私聊模式无DM配置(t *testing.T) {
	router := NewNotifyRouter(NotifyMode{Group: false, DM: true}, "test-robot", nil)
	d := router.BuildDispatcher([]string{"张三"})
	// 仅 Console（DM mode 开启但 dmHandler 为 nil）
	assert.Equal(t, 1, d.HandlerCount(), "私聊模式(无DM配置)应仅注册 Console")
}

func TestBuildDispatcher_全开模式(t *testing.T) {
	router := NewNotifyRouter(NotifyMode{Group: true, DM: true}, "test-robot", noopHandler{})
	d := router.BuildDispatcher([]string{"张三", "李四"})
	// Console + Card + DM = 3
	assert.Equal(t, 3, d.HandlerCount(), "全开模式应注册 Console + Card + DM")
}

func TestBuildDispatcher_全关模式(t *testing.T) {
	router := NewNotifyRouter(NotifyMode{Group: false, DM: false}, "test-robot", noopHandler{})
	d := router.BuildDispatcher(nil)
	// 仅 Console
	assert.Equal(t, 1, d.HandlerCount(), "全关模式应仅注册 Console")
}

func TestNewNotifyRouter_Mode返回值(t *testing.T) {
	mode := NotifyMode{Group: true, DM: false}
	router := NewNotifyRouter(mode, "test-robot", nil)
	assert.Equal(t, mode, router.Mode())
}

func TestBuildDispatcher_无作者群消息(t *testing.T) {
	router := NewNotifyRouter(NotifyMode{Group: true, DM: false}, "test-robot", nil)
	d := router.BuildDispatcher(nil)
	// 无作者也应正常注册 Console + Card
	assert.Equal(t, 2, d.HandlerCount(), "无作者时群消息模式应正常注册")
}

func TestBuildDispatcher_单作者群消息(t *testing.T) {
	router := NewNotifyRouter(NotifyMode{Group: true, DM: false}, "test-robot", nil)
	d := router.BuildDispatcher([]string{"张三"})
	assert.Equal(t, 2, d.HandlerCount(), "单作者应使用 WithAtUser")
}

func TestBuildDispatcher_多作者群消息(t *testing.T) {
	router := NewNotifyRouter(NotifyMode{Group: true, DM: false}, "test-robot", nil)
	d := router.BuildDispatcher([]string{"张三", "李四", "王五"})
	assert.Equal(t, 2, d.HandlerCount(), "多作者应使用 WithAtUsers")
}
