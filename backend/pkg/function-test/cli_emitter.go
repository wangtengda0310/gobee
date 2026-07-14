package functiontest

// cli_emitter.go 实现 CLI 场景下的机器人日志终端 emitter。
//
// 背景：RunRobotTest（services.go）通过 emitRobotLog 把每条 *log_service.Wails3Log
// 推给 emitter.Emit("robotLog", log, seq)。三个入口的 emitter 差异：
//   - GUI（cmd/rain-qa-func/wails.go）：传 app.Event，日志实时推前端 robotLog 事件；
//   - MCP（backend/pkg/settings/mcp/services.go）：传 nil，逐条日志丢弃，但
//     buildFightTestResult（mcp.go）会把全量 Logs 塞进返回值给 AI，故无需实时输出；
//   - CLI（cobra.go newFightTestCaseServiceForCLI）：此前传 nil，逐条 step 日志
//     被丢弃，终端只能看到 robot 库自带的粗粒度 stdout 日志，排查体验差。
//
// 本文件给 CLI 提供一个把日志格式化打印到 stdout 的 emitter，让 fight-test run
// 子命令能实时看到逐条执行过程（UseHeroSkill/PlayCard/OptRoomAction/断言结果…），
// 体验对齐前端 robot-test-log.vue 的渲染。
//
// 输出格式（对齐前端 robot-test-log.vue:179）：
//
//	[<Time>, 动作[<ID>], Step[<Case>], name[<RobotName>], [<Level>], <Msg>
//
// 注：前端 Step 字段优先用「用例 steps 列表里的 desc」，CLI 无此上下文，用 Case 名兜底。

import (
	"fmt"
	"io"
	"os"

	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/log_service"
)

// cliLogEmitter 把 robotLog 事件格式化打印到终端的 emitter。
// 实现 RobotLogEmitter 接口（见 emitter.go），供 CLI 入口（cobra.go）注入。
type cliLogEmitter struct {
	w io.Writer
}

// newCLILogEmitter 创建写入 os.Stdout 的终端 emitter，供 fight-test CLI 使用。
func newCLILogEmitter() *cliLogEmitter {
	return &cliLogEmitter{w: os.Stdout}
}

// Emit 实现 RobotLogEmitter 接口。
// 仅处理 "robotLog" 事件；data[0] 期望为 *log_service.Wails3Log，
// 类型不匹配或为 nil 时安全跳过（不 panic、不输出），返回 false。
// 其余事件名一律忽略，保证与 Wails EventManager 同构调用不会误输出。
//
// 并发安全：fmt.Fprintln 写 os.Stdout 底层走 os.File.Write（带互斥），
// RunRobotTest 的日志循环在 goroutine 中并发调用（services.go:222/226/245）安全。
func (e *cliLogEmitter) Emit(name string, data ...any) bool {
	if name != "robotLog" || len(data) == 0 {
		return false
	}
	log, ok := data[0].(*log_service.Wails3Log)
	if !ok || log == nil {
		return false
	}
	fmt.Fprintf(e.w, "[%s], 动作[%d], Step[%s], name[%s], [%s], %s\n",
		log.Time.Format("2006-01-02T15:04:05.000000"),
		log.ID,
		log.Case,
		log.RobotName,
		log.Level.String(),
		log.Msg,
	)
	return true
}
