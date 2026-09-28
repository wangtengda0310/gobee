package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

// 本文件的职责: 让所有路径定位不依赖"程序从哪个目录启动".
// 双击 exe / wails3 dev (go run, exe 在临时目录) / 从仓库根或任意子目录启动 / 未来独立分发, 均可命中.

// walkUp 从【可执行文件目录】与【当前工作目录】逐级向上, 用 probe 探测每级目录, 返回第一个非空结果.
// 先 exe 后 cwd: 生产环境以 exe 为准, 开发时 go run 的临时 exe 探测不中再靠 cwd 命中.
func walkUp(probe func(dir string) string) (string, error) {
	var starts []string
	if exe, err := os.Executable(); err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			exe = real
		}
		starts = append(starts, filepath.Dir(exe))
	}
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	for _, start := range starts {
		if start == "" {
			continue
		}
		for dir := start; ; {
			if hit := probe(dir); hit != "" {
				return hit, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break // 到达盘根
			}
			dir = parent
		}
	}
	return "", fmt.Errorf("定位失败 (已从 %v 逐级向上探测)", starts)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// bootScriptPath 定位 scripts\boot.ps1 (调度器入口).
// exe 在 bin\ / build\bin\ 等任意深度, 或 dev 时 cwd 在 winboot 项目根, 均可命中.
func bootScriptPath() (string, error) {
	// 优先 repo walk-up (开发特性: 改脚本无需重发 exe)
	script, err := walkUp(func(dir string) string {
		s := filepath.Join(dir, "scripts", "boot.ps1")
		if fileExists(s) {
			return s
		}
		return ""
	})
	if err == nil {
		return script, nil
	}
	// 独立运行兜底: 内置 scripts 解压到 %LOCALAPPDATA%\winboot\scripts
	if deployed, derr := ensureBundledScripts(); derr == nil {
		return deployed, nil
	}
	return "", fmt.Errorf("repo 内未找到 scripts/boot.ps1, 且内置脚本解压失败: %w", err)
}

// serverRootPath 定位 server\ 目录 (win_boot.ps1 / .server_id / compose 的根).
// 每级同时探测 <dir> 与 <dir>\server (后者覆盖"从仓库根启动"的场景).
// 兜底: WINBOOT_SERVER_DIR 环境变量显式指定 (exe 独立分发、仓库不在附近时).
func serverRootPath() (string, error) {
	if dir := os.Getenv("WINBOOT_SERVER_DIR"); dir != "" {
		if fileExists(filepath.Join(dir, "win_boot.ps1")) {
			return dir, nil
		}
		return "", fmt.Errorf("WINBOOT_SERVER_DIR=%q 指定的目录中没有 win_boot.ps1", dir)
	}
	return walkUp(func(dir string) string {
		for _, cand := range []string{dir, filepath.Join(dir, "server")} {
			if fileExists(filepath.Join(cand, "win_boot.ps1")) {
				return cand
			}
		}
		return ""
	})
}

// winbootLogsDir winboot/logs 目录 (PS 层 Start-Transcript 留档, /logs 报告用);
// 定位失败/不存在返回 "".
func winbootLogsDir() string {
	script, err := bootScriptPath()
	if err != nil {
		return ""
	}
	dir := filepath.Join(filepath.Dir(filepath.Dir(script)), "logs")
	if st, err := os.Stat(dir); err == nil && st.IsDir() {
		return dir
	}
	return ""
}

// ---------------------------------------------------------------------------
// 独立运行支持: scripts 经 go:embed 内置进 exe;
// repo 找不到脚本时解压到 %LOCALAPPDATA%\winboot\scripts (按版本刷新).
// ---------------------------------------------------------------------------

//go:embed all:scripts
var scriptsFS embed.FS

func winbootHome() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = os.Getenv("TEMP")
	}
	// 与 PyQt 参考版分开: 两栈版本号各自独立, 共用目录会互相覆盖解压脚本
	// 并触发对方的脚本版本门误报 (实测踩坑)
	return filepath.Join(base, "winboot-wails")
}

// scriptVersionOfPath 读 boot.ps1 的 $ScriptVersion 行 (缺失返回 "").
func scriptVersionOfPath(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return scriptVersionOfBytes(b)
}

func scriptVersionOfBytes(b []byte) string {
	m := scriptVersionRe.FindSubmatch(b)
	if m == nil {
		return ""
	}
	return string(m[1])
}

var scriptVersionRe = regexp.MustCompile(`(?m)^\s*\$ScriptVersion\s*=\s*'([^']+)'`)

// ensureBundledScripts 把内置 scripts 解压到 %LOCALAPPDATA%\winboot\scripts
// (版本不同才刷新), 返回部署后的 boot.ps1 路径.
func ensureBundledScripts() (string, error) {
	srcBoot, err := scriptsFS.ReadFile("scripts/boot.ps1")
	if err != nil {
		return "", err
	}
	srcVersion := scriptVersionOfBytes(srcBoot)
	deploy := filepath.Join(winbootHome(), "scripts")
	dstBoot := filepath.Join(deploy, "boot.ps1")
	if scriptVersionOfPath(dstBoot) != srcVersion {
		if err := os.MkdirAll(deploy, 0o755); err != nil {
			return "", err
		}
		err := fs.WalkDir(scriptsFS, "scripts", func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			data, err := scriptsFS.ReadFile(path)
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel("scripts", path)
			dst := filepath.Join(deploy, rel)
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return err
			}
			return os.WriteFile(dst, data, 0o644)
		})
		if err != nil {
			return "", err
		}
	}
	if _, err := os.Stat(dstBoot); err != nil {
		return "", err
	}
	return dstBoot, nil
}
