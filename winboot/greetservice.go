package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GreetService: server_id 的读取/持久化 (输入框"设置服务器名称"的后端).
// 路径经 serverRootPath() 定位 (paths.go), 与启动目录无关 ——
// 双击 exe / wails3 dev / 从任意 cwd 启动, 读写都落在 server\.server_id.
type GreetService struct{}

// server_id 校验规则与 win_boot.ps1 保持一致: 1-15 位小写字母/数字/连字符
var serverIDRe = regexp.MustCompile(`^[a-z0-9-]{1,15}$`)

// Greet 写入 server_id (保留模板方法名, 委托 SetServerID, 前端旧按钮不断)
func (g *GreetService) Greet(name string) string {
	return g.SetServerID(name)
}

// SetServerID 校验并写入 server\.server_id
func (g *GreetService) SetServerID(id string) string {
	id = strings.TrimSpace(id)
	if !serverIDRe.MatchString(id) {
		return "无效的 server_id: 必须是 1-15 位小写字母/数字/连字符 (当前: " + id + ")"
	}
	root, err := serverRootPath()
	if err != nil {
		return "定位 server 目录失败: " + err.Error()
	}
	if err := os.WriteFile(filepath.Join(root, ".server_id"), []byte(id), 0o644); err != nil {
		return "写入 .server_id 失败: " + err.Error()
	}
	return "已保存 server_id: " + id + " (位于 " + root + ")"
}

// GetServerID 读取 server\.server_id (不存在返回空串, 前端可用于预填输入框)
func (g *GreetService) GetServerID() string {
	root, err := serverRootPath()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(root, ".server_id"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
