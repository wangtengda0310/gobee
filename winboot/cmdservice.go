package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CmdService struct {
	mu        sync.Mutex
	logCmd    *exec.Cmd
	wantsLogs bool // 用户意图: 是否想跟随日志 (StopLogs 置 false, 断流自动重连的前提)
}

// lineWriter 把子进程输出按行拆开, 每行经事件推给前端 (GUI 实时日志的关键).
// 必须加锁: stdout/stderr 两根管道由 exec 的两个 goroutine 并发读取写入同一实例,
// 无锁时字节交错会把一行拦腰截断 (2026-09-24 实测: "命中本地缓存...DockerDesktopInst" 被截).
type lineWriter struct {
	prefix string
	mu     sync.Mutex
	buf    []byte
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		line := strings.TrimRight(string(w.buf[:i]), "\r")
		w.buf = w.buf[i+1:]
		if line != "" {
			// 经 hub.emitLog 汇入: 前端事件 + 环形缓冲 (/logs 报告数据源)
			emitLog(w.prefix + line)
		}
	}
	return len(p), nil
}

// Exec 执行 boot.ps1 子命令: doctor / install / uninstall / up / stop / restart 等.
// 输出实时逐行经 "boot:log" 事件推给前端 (带 [命令名] 前缀, 与容器日志同面板展示),
// 完整输出仍在返回值里 (供完成弹窗显示).
func (g *CmdService) Exec(name string, args []string) string {
	script, err := bootScriptPath()
	if err != nil {
		return "定位 boot.ps1 失败: " + err.Error()
	}
	// install 自动接入发现的局域网源 (acquire.ps1: 本地缓存 -> 局域网 -> 官方)
	if name == "install" {
		if src := hub.installLanBase(); src != "" {
			args = append(append([]string{}, args...), "-LanBase", src)
			emitLog(">>> 使用局域网源: " + src)
		}
	}
	full := append([]string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script, name}, args...)
	cmd := exec.Command("powershell.exe", full...)
	cmd.SysProcAttr = hideWindow() // 不闪黑窗 (跨平台经 build-tag 隔离)

	// stdout+stderr 双路: 实时推事件 + 攒完整输出
	var combined bytes.Buffer
	emitter := &lineWriter{prefix: "[" + name + "] "}
	cmd.Stdout = io.MultiWriter(&combined, emitter)
	cmd.Stderr = io.MultiWriter(&combined, emitter)

	if err := cmd.Start(); err != nil {
		return "启动命令失败: " + err.Error()
	}
	if err := cmd.Wait(); err != nil {
		code := cmd.ProcessState.ExitCode()
		hub.maybeBroadcastAnomaly(name, code, combined.String()) // 组内告警 (冷却窗口)
		return combined.String() + "\n[exit " + strconv.Itoa(code) + "] " + err.Error()
	}
	return combined.String()
}

// dockerPath 返回 docker.exe 路径 (PATH 优先, 兜底两种安装形态的已知位置)
func dockerPath() string {
	if p, err := exec.LookPath("docker"); err == nil {
		return p
	}
	candidates := []string{
		os.Getenv("ProgramFiles") + `\Docker\Docker\resources\bin\docker.exe`,
		os.Getenv("LOCALAPPDATA") + `\Programs\DockerDesktop\resources\bin\docker.exe`,
	}
	for _, bin := range candidates {
		if _, err := os.Stat(bin); err == nil {
			return bin
		}
	}
	return candidates[0]
}


// StartLogs 开始跟随容器日志.
// 关键行为: docker logs -f 在容器主进程退出(restart/stop)时会断流,
// 这里在 wantsLogs 仍为 true 时自动等待容器恢复运行并重连, 前端无需干预.
func (g *CmdService) StartLogs() string {
	g.mu.Lock()
	if g.wantsLogs {
		g.mu.Unlock()
		return "已在跟随日志"
	}
	g.wantsLogs = true
	g.mu.Unlock()

	go g.logLoop()
	return "ok"
}

// StopLogs 停止跟随日志 (只断流, 不影响容器)
func (g *CmdService) StopLogs() {
	g.mu.Lock()
	g.wantsLogs = false
	cmd := g.logCmd
	g.logCmd = nil
	g.mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
	}
}

// logLoop 跟随主循环: 等容器运行 -> attach 日志流 -> 断流后按意图重连.
// app 启动即跟随: 容器未运行时安静等待 (无期限), 容器起来后自动连接.
func (g *CmdService) logLoop() {
	for {
		if !g.waitRunning() {
			g.notifyLogEnd()
			return // 用户主动停止 (StopLogs)
		}
		emitLog("--- 已连接日志流 ---")
		if err := g.attach(); err != nil {
			emitLog("日志流异常: " + err.Error())
		}
		g.mu.Lock()
		want := g.wantsLogs
		g.mu.Unlock()
		if !want {
			g.notifyLogEnd()
			return // 用户主动停止
		}
		emitLog("--- 日志流断开 (容器重启/停止), 自动重连中... ---")
		time.Sleep(2 * time.Second)
	}
}

func (g *CmdService) notifyLogEnd() {
	if wailsApp != nil {
		wailsApp.Event.Emit("boot:logend", true)
	}
}

// attach 附着 docker logs -f, 流结束(容器停止/重启)时返回
func (g *CmdService) attach() error {
	cmd := exec.Command(dockerPath(), "logs", "-f", "--tail", "50", "sw_server_dev")
	cmd.SysProcAttr = hideWindow()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	g.mu.Lock()
	g.logCmd = cmd
	g.mu.Unlock()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 容错超长日志行(堆栈等)
	for scanner.Scan() {
		emitLog(scanner.Text())
	}
	return cmd.Wait()
}

// waitRunning 无限期轮询等待容器运行 (仅当用户停止跟随时返回 false).
// app 启动即跟随的场景: 容器/daemon 可能尚未就绪, 安静等待即可, 无需超时放弃.
func (g *CmdService) waitRunning() bool {
	waited := false
	for {
		g.mu.Lock()
		want := g.wantsLogs
		g.mu.Unlock()
		if !want {
			return false
		}
		cmd := exec.Command(dockerPath(), "inspect", "-f", "{{.State.Running}}", "sw_server_dev")
		cmd.SysProcAttr = hideWindow()
		out, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(out)) == "true" {
			return true
		}
		if !waited {
			emitLog("--- 容器未运行, 等待容器启动后自动连接... ---")
			waited = true
		}
		time.Sleep(5 * time.Second)
	}
}
