package pkg

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFindLatestVersion 测试findLatestVersion函数
func TestFindLatestVersion(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "version_test_*")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试结束后清理

	// 保存原始CommandDir
	originalCommandDir := CommandDir
	defer func() {
		// 测试结束后恢复原始CommandDir
		CommandDir = originalCommandDir
	}()

	// 设置测试用的CommandDir
	CommandDir = tempDir

	for _, suffix := range []string{".bat", "exe", "cmd"} {
		// 测试用例1: 命令目录下直接有可执行文件
		t.Run("ExecutableInCommandDir"+suffix, func(t *testing.T) {
			// 准备测试环境
			cmdName := "testcmd1"
			cmdDir := filepath.Join(tempDir, cmdName)
			err := os.MkdirAll(cmdDir, 0755)
			assert.NoError(t, err)

			// 创建可执行文件
			execPath := filepath.Join(cmdDir, cmdName)
			if runtime.GOOS == "windows" {
				execPath += suffix
			}
			err = os.WriteFile(execPath, []byte("dummy content"), 0755)
			assert.NoError(t, err)

			// 执行测试
			version, found, err, _ := findLatestVersion(cmdName)
			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, "", version)
		})
	}

	// 测试用例2: 存在latest目录且包含可执行文件
	t.Run("ExecutableInLatestDir", func(t *testing.T) {
		// 准备测试环境
		cmdName := "testcmd2"
		cmdDir := filepath.Join(tempDir, cmdName)
		latestDir := filepath.Join(cmdDir, "latest")
		err := os.MkdirAll(latestDir, 0755)
		assert.NoError(t, err)

		// 创建可执行文件
		execPath := filepath.Join(latestDir, cmdName)
		if runtime.GOOS == "windows" {
			execPath += ".exe"
		}
		err = os.WriteFile(execPath, []byte("dummy content"), 0755)
		assert.NoError(t, err)

		// 执行测试
		version, found, err, _ := findLatestVersion(cmdName)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "latest", version)
	})

	// 测试用例3: 从多个版本目录中选择最高版本
	t.Run("HighestVersionFromMultiple", func(t *testing.T) {
		// 准备测试环境
		cmdName := "testcmd3"
		cmdDir := filepath.Join(tempDir, cmdName)

		// 创建多个版本目录
		versions := []string{"1.0.0", "2.0.0", "0.9.0"}
		for _, v := range versions {
			versionDir := filepath.Join(cmdDir, v)
			err := os.MkdirAll(versionDir, 0755)
			assert.NoError(t, err)

			// 在每个版本目录中创建可执行文件
			execPath := filepath.Join(versionDir, cmdName)
			if runtime.GOOS == "windows" {
				execPath += ".exe"
			}
			err = os.WriteFile(execPath, []byte("dummy content"), 0755)
			assert.NoError(t, err)
		}

		// 执行测试
		version, found, err, _ := findLatestVersion(cmdName)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "2.0.0", version) // 应该返回最高版本
	})

	// 测试用例4: 没有有效版本目录
	t.Run("NoValidVersions", func(t *testing.T) {
		// 准备测试环境
		cmdName := "testcmd4"
		cmdDir := filepath.Join(tempDir, cmdName)
		err := os.MkdirAll(cmdDir, 0755)
		assert.NoError(t, err)

		// 创建一些无效的版本目录
		invalidDir := filepath.Join(cmdDir, "invalid-version")
		err = os.MkdirAll(invalidDir, 0755)
		assert.NoError(t, err)

		// 执行测试
		_, found, err, _ := findLatestVersion(cmdName)
		assert.Error(t, err) // 应该返回错误
		assert.False(t, found)
		assert.Contains(t, err.Error(), "没有有效的版本")
	})

	// 测试用例5: 命令目录不存在
	t.Run("CommandDirNotExist", func(t *testing.T) {
		// 使用不存在的命令名
		cmdName := "nonexistentcmd"

		// 执行测试
		_, found, err, _ := findLatestVersion(cmdName)
		assert.Error(t, err) // 应该返回错误
		assert.False(t, found)
	})
}
