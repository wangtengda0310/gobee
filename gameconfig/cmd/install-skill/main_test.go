// install-skill 回归测试
package main

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed skills/*
var testSkillFiles embed.FS

func TestSkillFilesIntegrity(t *testing.T) {
	// 验证嵌入的技能文件完整性
	requiredFiles := []string{
		"skills/SKILL.md",
		"skills/README.md",
		"skills/abilities/AI指导.md",
	}

	for _, file := range requiredFiles {
		// 尝试读取文件
		content, err := testSkillFiles.ReadFile(file)
		if err != nil {
			t.Errorf("缺少或无法读取文件: %s (%v)", file, err)
			continue
		}

		// 验证文件不为空
		if len(content) == 0 {
			t.Errorf("文件为空: %s", file)
		}

		// 验证 SKILL.md 包必需的内容
		if strings.HasSuffix(file, "SKILL.md") {
			contentStr := string(content)
			if !strings.Contains(contentStr, "gameconfig") {
				t.Errorf("SKILL.md 缺少预期内容")
			}
		}
	}
}

func TestSkillFileList(t *testing.T) {
	// 列出所有嵌入的文件
	foundFiles := make(map[string]bool)
	err := fs.WalkDir(testSkillFiles, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			foundFiles[path] = true
		}
		return nil
	})

	if err != nil {
		t.Fatalf("遍历文件失败: %v", err)
	}

	// 验证必需文件存在
	requiredFiles := []string{
		"skills/SKILL.md",
		"skills/README.md",
		"skills/abilities/AI指导.md",
	}

	for _, file := range requiredFiles {
		if !foundFiles[file] {
			t.Errorf("缺少必需文件: %s", file)
		}
	}

	// 打印找到的文件（调试用）
	t.Logf("找到 %d 个文件:", len(foundFiles))
	for file := range foundFiles {
		t.Logf("  - %s", file)
	}
}

func TestGetTargetDir_Default(t *testing.T) {
	// 测试默认目标目录
	targetDir = ""

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("无法获取用户主目录")
	}

	expected := filepath.Join(homeDir, ".claude", "skills", skillName)
	actual, err := getTargetDir()
	if err != nil {
		t.Fatalf("获取目标目录失败: %v", err)
	}

	if actual != expected {
		t.Errorf("目标目录不匹配\n期望: %s\n实际: %s", expected, actual)
	}
}

func TestGetTargetDir_Custom(t *testing.T) {
	// 测试自定义目标目录
	customDir := "/custom/skills/path"
	targetDir = customDir

	actual, err := getTargetDir()
	if err != nil {
		t.Fatalf("获取目标目录失败: %v", err)
	}

	if actual != customDir {
		t.Errorf("自定义目录不匹配\n期望: %s\n实际: %s", customDir, actual)
	}
}

func TestSkillFilesIntegrity(t *testing.T) {
	// 验证嵌入的技能文件完整性
	requiredFiles := []string{
		"skills/SKILL.md",
		"skills/README.md",
		"skills/abilities/AI指导.md",
	}

	for _, file := range requiredFiles {
		info, err := testSkillFiles.Stat(file)
		if err != nil {
			t.Errorf("缺少必需文件: %s (%v)", file, err)
			continue
		}

		if info.IsDir() {
			t.Errorf("路径是目录而非文件: %s", file)
			continue
		}

		// 验证文件不为空
		content, err := testSkillFiles.ReadFile(file)
		if err != nil {
			t.Errorf("读取文件失败: %s (%v)", file, err)
			continue
		}

		if len(content) == 0 {
			t.Errorf("文件为空: %s", file)
		}
	}
}

// TestInstallFlow 模拟完整安装流程
func TestInstallFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "skill-install-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 使用测试嵌入的文件系统
	oldSkillFiles := skillFiles
	skillFiles = testSkillFiles
	defer func() { skillFiles = oldSkillFiles }()

	// 设置自定义目标目录
	targetDir = tempDir

	// 执行安装流程
	if err := run(); err != nil {
		t.Fatalf("安装失败: %v", err)
	}

	// 验证关键文件存在
	keyFiles := []string{
		"SKILL.md",
		"abilities/AI指导.md",
	}

	for _, file := range keyFiles {
		path := filepath.Join(tempDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("安装后缺少文件: %s", file)
		}
	}
}

// TestSkillFileSync 检查技能文件是否与源文件同步
func TestSkillFileSync(t *testing.T) {
	// 读取嵌入的文件
	embeddedContent, err := testSkillFiles.ReadFile("skills/SKILL.md")
	if err != nil {
		t.Skip("无法读取嵌入文件，跳过同步检查")
	}

	// 读取源文件（如果存在）
	// 注意：这个测试需要在项目根目录运行
	sourcePath := filepath.Join("..", "..", ".claude", "skills", "gameconfig", "SKILL.md")
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Skip("无法读取源文件，跳过同步检查")
	}

	// 比较文件大小（简单检查）
	if len(embeddedContent) != len(sourceContent) {
		t.Errorf("技能文件可能未同步\n嵌入大小: %d\n源文件大小: %d\n请运行: cd cmd/install-skill && go generate",
			len(embeddedContent), len(sourceContent))
	}
}
