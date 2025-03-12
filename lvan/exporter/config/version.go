package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/wangtengda/gobee/lvan/exporter/logger"
	"golang.org/x/mod/semver"
)

var CommandDir string

// CommandInfo 存储命令的信息
type CommandInfo struct {
	Name       string // 命令名称
	Version    string // 版本号
	Executable string // 可执行文件路径
	IsLatest   bool   // 是否是最新版本
}

// GetCommandPath 获取命令的可执行文件路径
// cmdName: 命令名称
// version: 版本号，如果为"latest"则返回最新版本
// 返回值: 可执行文件路径, 是否找到, 错误信息
func GetCommandPath(cmdName, version string) (string, bool, error) {
	// 确保配置已初始化
	// 不需要传递参数，因为在main.go中已经初始化了配置

	// 检查命令目录是否存在
	cmdDir := filepath.Join(CommandDir, cmdName)
	if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
		return "", false, fmt.Errorf("命令 %s 不存在", cmdName)
	}

	// 如果请求未指定版本，查找最新版本
	if version == "" {
		latestVersion, found, err := findLatestVersion(cmdName)
		if err != nil || !found {
			return "", false, fmt.Errorf("找不到命令 %s 的最新版本: %v", cmdName, err)
		}
		version = latestVersion
		logger.Info("使用命令 %s 的最新版本: %s", cmdName, version)
	}

	// 构建版本目录路径
	versionDir := filepath.Join(cmdDir, version)
	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		logger.Info("命令 %s 的版本 %s 不存在", cmdName, version)
		return "", false, fmt.Errorf("命令 %s 的版本 %s 不存在", cmdName, version)
	}

	logger.Info("使用命令 %s 的版本 %s 目录 %s", cmdName, version, versionDir)

	// 查找可执行文件
	executable, found, err := findExecutable(versionDir, cmdName)
	if err != nil || !found {
		return "", false, fmt.Errorf("找不到命令 %s 版本 %s 的可执行文件: %v", cmdName, version, err)
	}

	return executable, true, nil
}

// findLatestVersion 查找命令的最新版本 跟目录 > latest目录 > sort max 目录
// 返回值: 最新版本号, 是否找到, 错误信息
func findLatestVersion(cmdName string) (string, bool, error) {
	cmdDir := filepath.Join(CommandDir, cmdName)

	// 检查命令目录下是否有可执行文件
	execPath := filepath.Join(cmdDir, cmdName)
	if isExecutable(execPath) {
		return "", true, nil
	}
	if isExecutable(execPath + ".exe") {
		return "", true, nil
	}
	if isExecutable(execPath + ".bat") {
		return "", true, nil
	}
	if isExecutable(execPath + ".cmd") {
		return "", true, nil
	}

	// 检查是否有latest目录并包含可执行文件
	latestDir := filepath.Join(cmdDir, "latest")
	if _, err := os.Stat(latestDir); !os.IsNotExist(err) {
		if _, found, err := findExecutable(latestDir, cmdName); found && err == nil {
			return "latest", true, nil
		}
	}

	// 读取版本目录
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return "", false, err
	}
	// 收集有效的版本号
	var versions []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 尝试解析版本号
		version := entry.Name()
		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}
		if !semver.IsValid(version) {
			// 忽略无效的版本号
			continue
		}

		versions = append(versions, version)
	}

	if len(versions) == 0 {
		return "", false, fmt.Errorf("命令 %s 没有有效的版本", cmdName)
	}

	// 按版本号排序
	// 使用 semver.Compare 函数对版本号进行排序
	sort.Slice(versions, func(i, j int) bool {
		return semver.Compare(versions[i], versions[j]) > 0
	})

	// 返回最高版本
	return strings.TrimPrefix(versions[0], "v"), true, nil
}

// findExecutable 在指定目录中查找可执行文件
// 返回值: 可执行文件路径, 是否找到, 错误信息
func findExecutable(dir, cmdName string) (string, bool, error) {
	// 首先尝试查找与命令名相同的可执行文件
	possibleNames := []string{
		cmdName, // Linux/macOS
	}

	// Windows 平台下，检查是否需要添加 .exe 后缀
	if runtime.GOOS == "windows" {
		logger.Info("当前windows环境")
		// 检查是否已有扩展名
		if !strings.Contains(filepath.Base(cmdName), ".") {
			// 检查 .exe 版本是否存在
			possibleNames = append(possibleNames, cmdName+".exe")

			// 检查 .bat 版本是否存在
			possibleNames = append(possibleNames, cmdName+".bat")

			// 检查 .cmd 版本是否存在
			possibleNames = append(possibleNames, cmdName+".cmd")
		}
	}

	for _, name := range possibleNames {
		path := filepath.Join(dir, name)
		if isExecutable(path) {
			return path, true, nil
		}
	}

	return "", false, fmt.Errorf("没有找到与命令名相同的可执行文件 %v", possibleNames)
}

// isExecutable 检查文件是否可执行
func isExecutable(path string) bool {
	// 检查文件是否存在
	info, err := os.Stat(path)
	if err != nil {
		logger.Info("没找到文件 %v", path)
		return false
	}

	// 在Windows上，检查文件扩展名是否为.exe
	if strings.HasSuffix(strings.ToLower(path), ".exe") || strings.HasSuffix(strings.ToLower(path), ".bat") || strings.HasSuffix(strings.ToLower(path), ".cmd") {
		return true
	}

	executable := info.Mode()&0111 != 0

	if !executable {
		logger.Info("文件 %v 没有执行权限", path)
		return false
	}
	// 在类Unix系统上，检查文件是否有执行权限
	return executable
}

// ListCommands 列出所有可用的命令及其版本
func ListCommands() ([]CommandInfo, error) {
	// 确保配置已初始化
	// 不需要传递参数，因为在main.go中已经初始化了配置

	var commands []CommandInfo

	// 读取命令目录
	cmdEntries, err := os.ReadDir(CommandDir)
	if err != nil {
		return nil, err
	}

	// 遍历每个命令
	for _, cmdEntry := range cmdEntries {
		if !cmdEntry.IsDir() {
			continue
		}

		cmdName := cmdEntry.Name()
		cmdDir := filepath.Join(CommandDir, cmdName)

		// 获取最新版本
		latestVersion, latestFound, _ := findLatestVersion(cmdName)

		// 读取版本目录
		versionEntries, err := os.ReadDir(cmdDir)
		if err != nil {
			continue
		}

		// 遍历每个版本
		for _, versionEntry := range versionEntries {
			if !versionEntry.IsDir() {
				continue
			}

			version := versionEntry.Name()
			versionDir := filepath.Join(cmdDir, version)

			// 查找可执行文件
			executable, found, _ := findExecutable(versionDir, cmdName)
			if !found {
				continue
			}

			// 添加到命令列表
			commands = append(commands, CommandInfo{
				Name:       cmdName,
				Version:    version,
				Executable: executable,
				IsLatest:   latestFound && version == latestVersion,
			})
		}
	}

	return commands, nil
}
