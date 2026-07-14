package functiontest

// cli_emitter_test.go 覆盖 CLI 终端 emitter（cli_emitter.go）的全部分支：
// 格式化输出、事件名过滤、nil 守卫、类型断言失败、空 data、Level 经 String()。
//
// 构造 emitter 时直接注入 *bytes.Buffer，避免依赖 os.Stdout，保证断言可重复。

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	log_def "git.devcloud.ztgame.com/v-tangfangda/rain-robot/log_def"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/log_service"
)

// newEmitterWithBuffer 构造一个写入 buffer 的 cliLogEmitter，便于断言输出。
func newEmitterWithBuffer() (*cliLogEmitter, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return &cliLogEmitter{w: buf}, buf
}

// sampleLog 构造一个字段固定的 *Wails3Log 供多数测试复用。
// Time 用 UTC 固定值，确保跨机器/时区输出确定。
func sampleLog() *log_service.Wails3Log {
	return &log_service.Wails3Log{
		Case:      "赵云_七进七出",
		ID:        3,
		Level:     log_def.INFO,
		RobotName: "pf_qax0x1",
		Time:      time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC),
		Msg:       "UseHeroSkill Start",
	}
}

func TestCLILogEmitter_格式化输出(t *testing.T) {
	emitter, buf := newEmitterWithBuffer()
	emitter.Emit("robotLog", sampleLog())

	want := "[2026-07-07T12:00:00.000000], 动作[3], Step[赵云_七进七出], name[pf_qax0x1], [INFO], UseHeroSkill Start\n"
	assert.Equal(t, want, buf.String())
}

func TestCLILogEmitter_非robotLog事件忽略(t *testing.T) {
	emitter, buf := newEmitterWithBuffer()
	emitter.Emit("otherEvent", sampleLog())
	assert.Empty(t, buf.String())
}

func TestCLILogEmitter_nil日志不panic(t *testing.T) {
	emitter, buf := newEmitterWithBuffer()
	assert.NotPanics(t, func() {
		emitter.Emit("robotLog", (*log_service.Wails3Log)(nil))
	})
	assert.Empty(t, buf.String())
}

func TestCLILogEmitter_类型断言失败安全(t *testing.T) {
	emitter, buf := newEmitterWithBuffer()
	assert.NotPanics(t, func() {
		emitter.Emit("robotLog", "not a log")
	})
	assert.Empty(t, buf.String())
}

func TestCLILogEmitter_空data安全(t *testing.T) {
	emitter, buf := newEmitterWithBuffer()
	assert.NotPanics(t, func() {
		emitter.Emit("robotLog")
	})
	assert.Empty(t, buf.String())
}

func TestCLILogEmitter_Level用String(t *testing.T) {
	emitter, buf := newEmitterWithBuffer()
	log := sampleLog()
	log.Level = log_def.WARN
	emitter.Emit("robotLog", log)

	// Level 应经 String() 输出 "WARN"，而非裸数字 "3"
	assert.Contains(t, buf.String(), "[WARN]")
}
