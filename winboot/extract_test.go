package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureBundledScripts(t *testing.T) {
	// 临时重定向 home: 避免污染真实部署目录
	t.Setenv("LOCALAPPDATA", t.TempDir())
	// 先验证嵌入内容可读
	src, err := scriptsFS.ReadFile("scripts/boot.ps1")
	if err != nil {
		t.Fatalf("嵌入脚本读取失败: %v", err)
	}
	if scriptVersionOfBytes(src) != TOOL_VERSION {
		t.Fatalf("嵌入脚本版本 %q != TOOL_VERSION %q", scriptVersionOfBytes(src), TOOL_VERSION)
	}
	p, err := ensureBundledScripts()
	if err != nil {
		t.Fatalf("解压失败: %v", err)
	}
	if !filepath.IsAbs(p) {
		t.Fatalf("非绝对路径: %s", p)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("解压产物缺失: %v", err)
	}
	if scriptVersionOfPath(p) != TOOL_VERSION {
		t.Fatalf("解压后版本不符")
	}
	t.Logf("解压到: %s", p)
}
