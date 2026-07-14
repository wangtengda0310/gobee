// Package appconfig 提供统一配置文件 .rain-qa-func.json 的 section 级读写。
//
// 各模块的配置（function-test、excel-test、hero-wiki-check、activity-wiki-check、mcp）
// 不再各自维护独立的 JSON 文件，而是统一存储在 .rain-qa-func.json 的不同 section 下。
//
// 用法：
//
//	cfg := appconfig.New("my_section")   // section 名
//	cfg.Load(&myConfig)                   // 读取
//	cfg.Save(&myConfig)                   // 写入（保留其他 section）
package appconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// ConfigFileName 统一配置文件名
const ConfigFileName = ".rain-qa-func.json"

var (
	globalMu       sync.Mutex
	globalFilePath string
)

// init 初始化配置文件绝对路径
func init() {
	globalFilePath = filepath.Join(configDir(), ConfigFileName)
}

// configDir 返回配置文件所在目录。
//
// 桌面端（windows/darwin/linux）沿用 CWD，保持向后兼容；
// Android 端 c-shared 进程的 CWD 为 "/" 且进程环境无 HOME/TMPDIR/XDG
// （os.UserConfigDir/UserHomeDir 均失败），若仍用 CWD，配置会落到只读的
// /.rain-qa-func.json 导致持久化失败，故改用应用私有可写目录 files/。
func configDir() string {
	if runtime.GOOS != "android" {
		if cwd, err := os.Getwd(); err == nil {
			return cwd
		}
	}
	return androidFilesDir()
}

// androidFilesDir 返回 Android 应用私有可写目录 /data/data/<package>/files。
// 包名优先从 /proc/self/cmdline 读取（应用进程首字段），读取失败时回退到
// 构建期包名 com.wails.app（见 build/android/app/build.gradle applicationId）。
func androidFilesDir() string {
	if data, err := os.ReadFile("/proc/self/cmdline"); err == nil {
		parts := strings.Split(strings.TrimRight(string(data), "\x00"), "\x00")
		if len(parts) > 0 && parts[0] != "" {
			return filepath.Join("/data/data", parts[0], "files")
		}
	}
	return "/data/data/com.wails.app/files"
}

// FilePath 返回统一配置文件的绝对路径
func FilePath() string {
	globalMu.Lock()
	defer globalMu.Unlock()
	return globalFilePath
}

// Section 针对单个 section 的读写器
type Section struct {
	name string
}

// New 创建一个 section 读写器
func New(name string) *Section {
	return &Section{name: name}
}

// readAll 读取整个配置文件（文件不存在时返回空 map）
func readAll() (map[string]json.RawMessage, error) {
	path := FilePath()
	data := make(map[string]json.RawMessage)

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return data, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}
	if len(raw) == 0 {
		return data, nil
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}
	return data, nil
}

// writeAll 写入整个配置文件（格式化输出）
func writeAll(data map[string]json.RawMessage) error {
	path := FilePath()
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	if err := os.WriteFile(path, out, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}
	return nil
}

// Load 从统一配置文件的对应 section 读取到 v
// section 不存在时返回 nil（不报错），由调用方使用默认值
func (s *Section) Load(v any) error {
	data, err := readAll()
	if err != nil {
		return err
	}
	section, ok := data[s.name]
	if !ok {
		return nil
	}
	return json.Unmarshal(section, v)
}

// Exists 检查 section 是否存在
func (s *Section) Exists() bool {
	data, err := readAll()
	if err != nil {
		return false
	}
	_, ok := data[s.name]
	return ok
}

// Save 将 v 序列化后写入对应 section，保留其他 section 不变
func (s *Section) Save(v any) error {
	data, err := readAll()
	if err != nil {
		return err
	}
	sectionData, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("序列化 section 失败: %v", err)
	}
	data[s.name] = sectionData
	return writeAll(data)
}

// Delete 删除对应 section，保留其他 section 不变
func (s *Section) Delete() error {
	data, err := readAll()
	if err != nil {
		return err
	}
	delete(data, s.name)
	return writeAll(data)
}
