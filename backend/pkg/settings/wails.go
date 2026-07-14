package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/appconfig"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu"
)

// ========== ClaudeCodeSettings - Claude Code 配置 ==========

// ClaudeCodeSettings Claude Code 配置文件结构
// 对应 ~/.claude/settings.json
type ClaudeCodeSettings struct {
	Env map[string]string `json:"env"`
}

// LoadClaudeCodeConfig 加载 Claude Code 配置文件
// 返回配置对象和可能的错误
func LoadClaudeCodeConfig() (*ClaudeCodeSettings, error) {
	// 获取用户主目录
	var homeDir string
	if runtime.GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}

	if homeDir == "" {
		return nil, fmt.Errorf("无法获取用户主目录")
	}

	// Claude Code 配置文件路径
	configPath := filepath.Join(homeDir, ".claude", "settings.json")

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取 Claude Code 配置失败: %w", err)
	}

	// 解析 JSON
	var settings ClaudeCodeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("解析 Claude Code 配置失败: %w", err)
	}

	return &settings, nil
}

// GetClaudeCodeConfigInfo 获取 Claude Code 配置信息（用于前端展示）
// 返回是否可用以及配置的模型名称
func GetClaudeCodeConfigInfo() (available bool, model string, err error) {
	settings, err := LoadClaudeCodeConfig()
	if err != nil {
		return false, "", err
	}

	model = settings.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"]
	if model == "" {
		model = settings.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"]
	}
	if model == "" {
		model = settings.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"]
	}

	return true, model, nil
}

// ========== FeishuNotifyConfig - 飞书通知配置 ==========

// FeishuNotifyConfig 飞书自动通知配置
// [MCP改进] 支持测试完成后自动发送飞书通知
// 使用 common/feishu 包中的类型
type FeishuNotifyConfig = feishu.FeishuNotifyConfig

// DefaultFeishuMessageTemplate 默认消息模板
const DefaultFeishuMessageTemplate = feishu.DefaultFeishuMessageTemplate

// FeishuNotifyConfigService 飞书通知配置服务
type FeishuNotifyConfigService struct {
	configFile string
}

// NewFeishuNotifyConfigService 创建飞书通知配置服务实例
func NewFeishuNotifyConfigService() *FeishuNotifyConfigService {
	// 配置文件放在当前目录下的 feishu_notify_config.json
	configFile := "feishu_notify_config.json"

	// 如果是相对路径，转换为绝对路径
	if !filepath.IsAbs(configFile) {
		cwd, err := os.Getwd()
		if err == nil {
			configFile = filepath.Join(cwd, configFile)
		}
	}

	return &FeishuNotifyConfigService{
		configFile: configFile,
	}
}

// getDefaultConfig 获取默认配置
func (s *FeishuNotifyConfigService) getDefaultConfig() *FeishuNotifyConfig {
	return &FeishuNotifyConfig{
		Enabled:         false,
		RobotGUID:       "",
		MessageTemplate: DefaultFeishuMessageTemplate,
	}
}

// GetConfig 获取当前配置
func (s *FeishuNotifyConfigService) GetConfig() (*FeishuNotifyConfig, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(s.configFile); os.IsNotExist(err) {
		// 文件不存在，创建默认配置并保存
		defaultConfig := s.getDefaultConfig()
		if err := s.SaveConfig(defaultConfig); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %v", err)
		}
		return defaultConfig, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(s.configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config FeishuNotifyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 如果消息模板为空，使用默认模板
	if config.MessageTemplate == "" {
		config.MessageTemplate = DefaultFeishuMessageTemplate
	}

	return &config, nil
}

// SaveConfig 保存配置
func (s *FeishuNotifyConfigService) SaveConfig(config *FeishuNotifyConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 如果消息模板为空，使用默认模板
	if config.MessageTemplate == "" {
		config.MessageTemplate = DefaultFeishuMessageTemplate
	}

	// 序列化为 JSON，格式化输出
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(s.configFile, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// UpdateConfig 更新配置（部分更新）
func (s *FeishuNotifyConfigService) UpdateConfig(updates map[string]interface{}) (*FeishuNotifyConfig, error) {
	// 获取当前配置
	currentConfig, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	// 临时转换为 map 以便更新
	configMap := make(map[string]interface{})
	data, _ := json.Marshal(currentConfig)
	json.Unmarshal(data, &configMap)

	// 应用更新
	for key, value := range updates {
		configMap[key] = value
	}

	// 转换回 FeishuNotifyConfig 结构体
	updatedData, _ := json.Marshal(configMap)
	var updatedConfig FeishuNotifyConfig
	if err := json.Unmarshal(updatedData, &updatedConfig); err != nil {
		return nil, fmt.Errorf("更新配置失败: %v", err)
	}

	// 保存更新后的配置
	if err := s.SaveConfig(&updatedConfig); err != nil {
		return nil, err
	}

	return &updatedConfig, nil
}

// ResetToDefault 重置为默认配置
func (s *FeishuNotifyConfigService) ResetToDefault() (*FeishuNotifyConfig, error) {
	defaultConfig := s.getDefaultConfig()
	if err := s.SaveConfig(defaultConfig); err != nil {
		return nil, err
	}
	return defaultConfig, nil
}

// SendNotification 发送飞书通知（实现 feishu.FeishuNotifier 接口）
func (s *FeishuNotifyConfigService) SendNotification(heroName string, total, passed, failed int, passRate float64) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}

	if !config.Enabled {
		return nil
	}

	if config.RobotGUID == "" {
		return nil
	}

	// 使用 feishu 包中的函数构建消息
	message := buildFeishuMessage(config.MessageTemplate, heroName, total, passed, failed, passRate)

	// 发送飞书消息
	feishu.SendFeiShuRobotText(config.RobotGUID, "%s", message)

	return nil
}

// buildFeishuMessage 构建飞书消息（复制自 feishu 包）
func buildFeishuMessage(template, heroName string, total, passed, failed int, passRate float64) string {
	message := template
	message = strings.ReplaceAll(message, "{heroName}", heroName)
	message = strings.ReplaceAll(message, "{total}", fmt.Sprintf("%d", total))
	message = strings.ReplaceAll(message, "{passed}", fmt.Sprintf("%d", passed))
	message = strings.ReplaceAll(message, "{failed}", fmt.Sprintf("%d", failed))
	message = strings.ReplaceAll(message, "{passRate}", fmt.Sprintf("%.1f", passRate))
	return message
}

// ========== MCPConfig - MCP 配置 ==========

// MCPConfig MCP 配置结构（用于前端绑定）
type MCPConfig struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port"`
	Host    string `json:"host"`
}

// Address 返回服务器地址
func (c *MCPConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// MCPServerInterface MCP 服务器接口（在 internal 包中定义，避免循环导入）
// mcp.MCPServer 会隐式实现这个接口
type MCPServerInterface interface {
	GetConfig() *MCPConfig
	IsRunning() bool
	GetConnectionCount() int
	Stop(ctx context.Context) error
	Restart(ctx context.Context, newConfig *MCPConfig) error
	UpdateConfigAndRestart(enabled bool, port int, host string) error
}

// MCPConfigService MCP 配置管理服务
type MCPConfigService struct {
	server MCPServerInterface
	mu     sync.RWMutex
}

// NewMCPConfigService 创建 MCP 配置服务实例
func NewMCPConfigService() *MCPConfigService {
	return &MCPConfigService{}
}

// SetServer 设置 MCP 服务器
func (s *MCPConfigService) SetServer(server MCPServerInterface) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server = server
}

// GetConfig 获取当前配置
// @frontend @mcp
func (s *MCPConfigService) GetConfig() (*MCPConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.server == nil {
		return &MCPConfig{
			Enabled: true,
			Port:    8765,
			Host:    "127.0.0.1",
		}, nil
	}

	config := s.server.GetConfig()
	return &MCPConfig{
		Enabled: config.Enabled,
		Port:    config.Port,
		Host:    config.Host,
	}, nil
}

// SaveConfig 保存配置
// @frontend @mcp
func (s *MCPConfigService) SaveConfig(config *MCPConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server == nil {
		return fmt.Errorf("MCP 服务器未初始化")
	}

	return s.server.UpdateConfigAndRestart(config.Enabled, config.Port, config.Host)
}

// StartMCPService 启动 MCP 服务
func (s *MCPConfigService) StartMCPService(enabled bool, port int, host string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server == nil {
		return fmt.Errorf("MCP 服务器未初始化")
	}

	ctx := context.Background()
	return s.server.Restart(ctx, &MCPConfig{
		Enabled: enabled,
		Port:    port,
		Host:    host,
	})
}

// StopMCPService 停止 MCP 服务
func (s *MCPConfigService) StopMCPService() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server == nil {
		return fmt.Errorf("MCP 服务器未初始化")
	}

	ctx := context.Background()
	return s.server.Stop(ctx)
}

// IsRunning 检查 MCP 服务是否在运行
// @frontend
func (s *MCPConfigService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.server == nil {
		return false
	}

	return s.server.IsRunning()
}

// GetConnectionCount 获取当前连接数
func (s *MCPConfigService) GetConnectionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.server == nil {
		return 0
	}

	return s.server.GetConnectionCount()
}

// GetServerAddress 获取服务器完整地址
func (s *MCPConfigService) GetServerAddress() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.server == nil {
		return "127.0.0.1:8765"
	}

	config := s.server.GetConfig()
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

// GetSSEEndpoint 获取 SSE 端点地址
func (s *MCPConfigService) GetSSEEndpoint() string {
	return fmt.Sprintf("http://%s/sse", s.GetServerAddress())
}

// GetMCPStatus 获取 MCP 服务状态
// @mcp
func (s *MCPConfigService) GetMCPStatus() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]any{
		"running":  false,
		"address":  "127.0.0.1:8765",
		"enabled":  true,
		"endpoint": "",
	}

	if s.server != nil {
		config := s.server.GetConfig()
		status["running"] = s.server.IsRunning()
		status["enabled"] = config.Enabled
		status["address"] = fmt.Sprintf("%s:%d", config.Host, config.Port)
		status["endpoint"] = fmt.Sprintf("http://%s:%d", config.Host, config.Port)
	}

	log.Printf("GetMCPStatus: %+v", status)
	return status
}

// GetConfigFilePath 获取配置文件路径
func (s *MCPConfigService) GetConfigFilePath() string {
	return appconfig.FilePath()
}

// ========== VersionService - 版本信息 ==========

// 构建时通过 -ldflags 注入的变量
// 生产构建时会通过 Taskfile.yml 注入实际的 git 信息
// 开发模式下这些变量为空字符串，会降级到运行时 git 命令获取
var (
	// VersionCommitHash 提交短哈希，如 "8de8a4b"
	VersionCommitHash string
	// VersionCommitMessage 提交消息，如 "feat: 添加状态栏"
	VersionCommitMessage string
	// VersionCommitAuthor 提交作者，如 "wangtengda"
	VersionCommitAuthor string
	// VersionCommitDate 提交相对时间，如 "2 hours ago"
	VersionCommitDate string
	// VersionBuildTime 构建时间，如 "2026-03-19 11:30:00"
	VersionBuildTime string
)

// CommitInfo 提交信息结构
type CommitInfo struct {
	Hash    string `json:"hash"`    // 短哈希
	Message string `json:"message"` // 提交消息
	Author  string `json:"author"`  // 作者
	Date    string `json:"date"`    // 相对日期
}

// BuildInfo 构建信息结构
type BuildInfo struct {
	CommitHash string `json:"commitHash"` // 提交哈希
	CommitMsg  string `json:"commitMsg"`  // 提交消息
	BuildTime  string `json:"buildTime"`  // 构建时间
}

// VersionService 版本信息服务
// 提供获取 Git 提交信息等版本相关功能
// 支持两种模式：
//   - 开发模式：运行时执行 git log 命令获取实时信息
//   - 生产构建：使用构建时注入的版本信息
type VersionService struct{}

// NewVersionService 创建版本服务实例
func NewVersionService() *VersionService {
	return &VersionService{}
}

// GetRecentCommits 获取最近 N 条 commit 信息
// 优先使用构建时注入的信息，降级到运行时 git 命令
// count 指定获取的提交数量，默认5条，最大20条
// @frontend
func (v *VersionService) GetRecentCommits(count int) []CommitInfo {
	if count <= 0 || count > 20 {
		count = 5
	}

	// 检查是否有构建时注入的版本信息
	if VersionCommitHash != "" {
		return v.getInjectedCommits(count)
	}

	// 降级：运行时获取
	return v.getGitCommits(count)
}

// GetBuildInfo 获取构建信息
// 返回构建时注入的版本信息，用于显示构建时间等
// @frontend
func (v *VersionService) GetBuildInfo() BuildInfo {
	return BuildInfo{
		CommitHash: VersionCommitHash,
		CommitMsg:  VersionCommitMessage,
		BuildTime:  VersionBuildTime,
	}
}

// GetCurrentDirectory 获取当前工作目录
// 返回程序启动时的当前工作目录，用于状态栏显示
// @frontend
func (v *VersionService) GetCurrentDirectory() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "未知目录"
	}
	return cwd
}

// getInjectedCommits 返回构建时注入的版本信息
// 如果需要更多历史记录，从 git 获取补充
func (v *VersionService) getInjectedCommits(count int) []CommitInfo {
	commits := []CommitInfo{{
		Hash:    VersionCommitHash,
		Message: VersionCommitMessage,
		Author:  VersionCommitAuthor,
		Date:    VersionCommitDate,
	}}

	// 如果需要更多历史记录，从 git 获取补充
	if count > 1 {
		gitCommits := v.getGitCommits(count)
		// 过滤掉与注入版本相同的 commit，并添加其他历史记录
		for _, c := range gitCommits {
			if c.Hash != VersionCommitHash {
				commits = append(commits, c)
			}
			if len(commits) >= count {
				break
			}
		}
	}

	return commits
}

// getGitCommits 运行时通过 git 命令获取提交信息
func (v *VersionService) getGitCommits(count int) []CommitInfo {
	workDir := v.findGitRoot()
	if workDir == "" {
		return []CommitInfo{{Hash: "N/A", Message: "未找到 Git 仓库", Author: "", Date: ""}}
	}

	// 使用 git log 命令获取提交信息
	// 格式: 短哈希|提交消息|作者|相对时间
	cmd := exec.Command("git", "log", "-n", strconv.Itoa(count),
		"--pretty=format:%h|%s|%an|%cr")
	cmd.Dir = workDir

	output, err := cmd.Output()
	if err != nil {
		return []CommitInfo{{Hash: "N/A", Message: "无法获取 Git 信息: " + err.Error(), Author: "", Date: ""}}
	}

	return parseCommitOutput(string(output))
}

// findGitRoot 查找 git 仓库根目录
func (v *VersionService) findGitRoot() string {
	// 获取当前可执行文件路径
	exePath, err := exec.LookPath(os.Args[0])
	if err != nil {
		exePath, _ = filepath.Abs(".")
	}

	// 从当前目录向上查找 .git 目录
	dir := filepath.Dir(exePath)
	for {
		gitDir := filepath.Join(dir, ".git")
		if _, err := exec.Command("git", "-C", dir, "rev-parse", "--git-dir").Output(); err == nil {
			// 检查 .git 是否存在
			if _, statErr := filepath.Abs(gitDir); statErr == nil {
				return dir
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// 已到达根目录，未找到
			return ""
		}
		dir = parent
	}
}

var osArgs = os.Args

// parseCommitOutput 解析 git log 输出
func parseCommitOutput(output string) []CommitInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var commits []CommitInfo

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) >= 2 {
			commit := CommitInfo{
				Hash:    parts[0],
				Message: parts[1],
			}
			if len(parts) >= 3 {
				commit.Author = parts[2]
			}
			if len(parts) >= 4 {
				commit.Date = parts[3]
			}
			commits = append(commits, commit)
		}
	}

	if len(commits) == 0 {
		commits = append(commits, CommitInfo{Hash: "N/A", Message: "无提交记录", Author: "", Date: ""})
	}

	return commits
}
