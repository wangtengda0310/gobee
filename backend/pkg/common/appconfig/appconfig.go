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
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ConfigFileName 统一配置文件名
const ConfigFileName = ".rain-qa-func.json"

var (
	globalMu       sync.Mutex
	globalFilePath string
)

// init 初始化配置文件绝对路径。
//
// 从当前工作目录逐级向上查找已存在的 .rain-qa-func.json，命中即用其路径；
// 一路未找到则回退到 cwd 下的默认路径（保持历史行为，首次 Save 在 cwd 创建）。
//
// 这样在仓库子目录下运行时仍能读到仓库根的统一配置；Save 写回找到的文件，
// 避免在子目录意外新建空配置。
func init() {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	path, result := findConfigFile(cwd)
	globalFilePath = path

	// 输出配置文件搜索过程，便于排查「为何读到了/没读到某个配置」
	switch result.found {
	case true:
		extra := ""
		if result.viaWorktree {
			extra = fmt.Sprintf("（经 worktree gitdir 跳转主仓库 %s）", result.worktreeMainRepo)
		}
		log.Printf("[appconfig] 配置文件: %s%s", path, extra)
	case false:
		log.Printf("[appconfig] 未找到 %s，回退默认路径 %s（搜索: %s）",
			ConfigFileName, path, result.searchedDirs)
	}
}

// configSearchResult 记录配置文件搜索过程，用于日志输出。
type configSearchResult struct {
	found            bool   // 是否找到已存在的配置文件
	viaWorktree      bool   // 是否经由 worktree gitdir 跳转命中
	worktreeMainRepo string // 跳转到的主仓库根（viaWorktree=true 时有效）
	searchedDirs     string // 未命中时记录的搜索轨迹（dir1 -> dir2 -> ...）
}

// findConfigFile 从 startDir 逐级向上查找 ConfigFileName，
// 返回首个命中的文件路径与搜索过程；未找到则返回 startDir 下的默认路径。
//
// 兼容 git worktree：当向上遇到 .git 为文件（而非目录）时，它是 worktree 的
// gitdir 指针（内容形如 "gitdir: <主仓库>/.git/worktrees/<name>"），据此跳转
// 到主仓库根继续查找——worktree 与主仓库可能位于不同盘符/独立目录树，单纯向上
// 遍历无法到达。
func findConfigFile(startDir string) (string, configSearchResult) {
	var searched []string
	dir := startDir
	for {
		if candidate := configInDir(dir); candidate != "" {
			return candidate, configSearchResult{found: true}
		}
		searched = append(searched, dir)
		// worktree：.git 是文件时跳转主仓库根
		if mainRepo := worktreeMainRepo(dir); mainRepo != "" {
			if candidate := configInDir(mainRepo); candidate != "" {
				return candidate, configSearchResult{
					found:            true,
					viaWorktree:      true,
					worktreeMainRepo: mainRepo,
				}
			}
			searched = append(searched, mainRepo+"(worktree主仓库)")
		}
		parent := filepath.Dir(dir)
		if parent == dir { // 已到文件系统根
			return filepath.Join(startDir, ConfigFileName), configSearchResult{
				found:        false,
				searchedDirs: strings.Join(searched, " -> "),
			}
		}
		dir = parent
	}
}

// configInDir 返回 dir 下存在的配置文件路径，不存在则返回空串。
func configInDir(dir string) string {
	candidate := filepath.Join(dir, ConfigFileName)
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		return candidate
	}
	return ""
}

// worktreeMainRepo 检查 dir/.git 是否为 worktree 的 gitdir 指针文件，
// 是则返回主仓库根目录，否则返回空串。
func worktreeMainRepo(dir string) string {
	gitPath := filepath.Join(dir, ".git")
	info, err := os.Stat(gitPath)
	if err != nil || info.IsDir() {
		return ""
	}
	raw, err := os.ReadFile(gitPath)
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(string(raw))
	const prefix = "gitdir:"
	if !strings.HasPrefix(line, prefix) {
		return ""
	}
	gitdir := strings.TrimSpace(line[len(prefix):])
	// worktree gitdir 形如 <主仓库>/.git/worktrees/<name>，取 .git 之前的部分。
	// 统一用正斜杠定位 / .git / 段，返回时再归一化为系统原生路径分隔符。
	normalized := filepath.ToSlash(gitdir)
	idx := strings.Index(normalized, "/.git/")
	if idx <= 0 {
		return ""
	}
	return filepath.FromSlash(normalized[:idx])
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
