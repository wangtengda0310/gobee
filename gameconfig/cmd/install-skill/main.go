// install-skill - 将 gameconfig skill 安装到全局 Claude Code skills 目录
//
// 使用:
//   go install github.com/wangtengda0310/gobee/gameconfig/cmd/install-skill@latest
//   gameconfig-install-skill
//
// 或在 gameconfig 仓库中:
//   go run cmd/install-skill/main.go

package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:generate cp -r ../../.claude/skills/gameconfig ./skills
//go:embed skills/*
var skillFiles embed.FS

const (
	skillName = "gameconfig"
	version   = "1.0.0"
)

var (
	targetDir  string
	showHelp   bool
	showVersion bool
)

func init() {
	flag.StringVar(&targetDir, "target", "", "自定义安装目标目录")
	flag.BoolVar(&showHelp, "help", false, "显示帮助")
	flag.BoolVar(&showVersion, "version", false, "显示版本")
}

func main() {
	flag.Parse()

	if showHelp {
		printHelp()
		return
	}

	if showVersion {
		fmt.Printf("gameconfig install-skill v%s\n", version)
		return
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 安装失败: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. 确定目标目录
	targetPath, err := getTargetDir()
	if err != nil {
		return fmt.Errorf("无法确定目标目录: %w", err)
	}

	// 2. 创建目标目录
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 3. 从嵌入的文件系统复制 skill 文件
	fmt.Printf("📦 正在安装 gameconfig skill...\n")
	fmt.Printf("   目标: %s\n", targetPath)

	if err := copyEmbeddedFiles(targetPath); err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	// 4. 成功
	fmt.Printf("✅ gameconfig skill 已安装到: %s\n", targetPath)
	fmt.Printf("\n现在你可以在任何项目中使用 gameconfig skill 了！\n")
	fmt.Printf("更多信息: https://github.com/wangtengda0310/gobee/gameconfig\n")

	return nil
}

// copyEmbeddedFiles 从嵌入的文件系统复制文件到目标目录
func copyEmbeddedFiles(targetPath string) error {
	// 遍历 skills/ 目录下的所有文件
	return fs.WalkDir(skillFiles, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 去掉 "skills/" 前缀
		filePath := path
		if len(path) > len("skills/") {
			filePath = path[len("skills/"):]
		} else if path == "skills" {
			// 跳过根目录
			return nil
		}

		dstPath := filepath.Join(targetPath, filePath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		// 读取嵌入的文件内容
		content, err := skillFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取嵌入文件失败 %s: %w", path, err)
		}

		// 写入目标文件
		return os.WriteFile(dstPath, content, 0644)
	})
}

// getTargetDir 获取目标安装目录
func getTargetDir() (string, error) {
	if targetDir != "" {
		return targetDir, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".claude", "skills", skillName), nil
}

func printHelp() {
	fmt.Printf(`gameconfig skill 安装工具 v%s

用法:
  go install github.com/wangtengda0310/gobee/gameconfig/cmd/install-skill@latest
  gameconfig-install-skill [选项]

或在 gameconfig 仓库中:
  go run cmd/install-skill/main.go [选项]

选项:
  -target <目录>   自定义安装目标目录
  -help           显示此帮助
  -version        显示版本

说明:
  将 gameconfig skill 安装到全局 Claude Code skills 目录。
  安装后，你可以在任何项目中使用 gameconfig 相关的 AI 能力。

  此工具已内嵌 skill 文件，无需 clone gameconfig 仓库即可使用。

默认安装位置: ~/.claude/skills/gameconfig/

更多信息: https://github.com/wangtengda0310/gobee/gameconfig
`, version)
}
