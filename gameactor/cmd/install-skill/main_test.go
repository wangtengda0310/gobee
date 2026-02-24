package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetTargetDir(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("无法获取用户主目录: %v", err)
	}

	expected := filepath.Join(homeDir, ".claude", "skills", "gameactor")
	actual, err := getTargetDir()
	if err != nil {
		t.Fatalf("getTargetDir 失败: %v", err)
	}

	if actual != expected {
		t.Errorf("期望 %s, 实际 %s", expected, actual)
	}
}

func TestGetTargetDir_Custom(t *testing.T) {
	customDir := "/tmp/custom/skills"
	targetDir = customDir

	actual, err := getTargetDir()
	if err != nil {
		t.Fatalf("getTargetDir 失败: %v", err)
	}

	if actual != customDir {
		t.Errorf("期望 %s, 实际 %s", customDir, actual)
	}

	// 重置
	targetDir = ""
}
