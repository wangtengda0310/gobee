package roadmap

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"gopkg.in/yaml.v3"
)

// ========== 常量定义 ==========

const (
	// MaxCommentLength 评论最大长度（字符数）
	MaxCommentLength = 2000
	// MaxTitleLength 标题最大长度
	MaxTitleLength = 50
	// MaxDescriptionLength 描述最大长度
	MaxDescriptionLength = 5000
	// ConfigFileName 配置文件名
	ConfigFileName = "roadmap.yaml"
)

// ========== 数据结构定义 ==========

// RoadmapStatus 功能状态
type RoadmapStatus string

const (
	StatusPlanning   RoadmapStatus = "planning"    // 规划中
	StatusInProgress RoadmapStatus = "in_progress" // 开发中
	StatusCompleted  RoadmapStatus = "completed"   // 已完成
	StatusRejected   RoadmapStatus = "rejected"    // 已拒绝
)

// Priority 优先级
type Priority string

const (
	PriorityLow    Priority = "low"    // 低
	PriorityMedium Priority = "medium" // 中
	PriorityHigh   Priority = "high"   // 高
)

// Votes 投票信息
type Votes struct {
	Up       int     `json:"up" yaml:"up"`                                   // 支持数
	Down     int     `json:"down" yaml:"down"`                               // 反对数
	UserVote *string `json:"user_vote,omitempty" yaml:"user_vote,omitempty"` // 当前用户投票
}

// Comment 评论
type Comment struct {
	ID        string `json:"id" yaml:"id"`                 // 评论ID
	Author    string `json:"author" yaml:"author"`         // 作者
	Content   string `json:"content" yaml:"content"`       // 内容
	CreatedAt int64  `json:"created_at" yaml:"created_at"` // 创建时间（Unix时间戳）
}

// RoadmapItem 路线图项目
type RoadmapItem struct {
	ID          string        `json:"id" yaml:"id"`                   // 项目ID
	Title       string        `json:"title" yaml:"title"`             // 标题
	Description string        `json:"description" yaml:"description"` // 描述
	Status      RoadmapStatus `json:"status" yaml:"status"`           // 状态
	Priority    Priority      `json:"priority" yaml:"priority"`       // 优先级
	Author      string        `json:"author" yaml:"author"`           // 创建者
	CreatedAt   int64         `json:"created_at" yaml:"created_at"`   // 创建时间
	UpdatedAt   int64         `json:"updated_at" yaml:"updated_at"`   // 更新时间
	Votes       Votes         `json:"votes" yaml:"votes"`             // 投票
	Comments    []Comment     `json:"comments" yaml:"comments"`       // 评论
}

// RoadmapConfig 路线图配置文件结构
type RoadmapConfig struct {
	Version string        `yaml:"version" json:"version"` // 版本号
	Items   []RoadmapItem `yaml:"items" json:"items"`     // 功能列表
}

// ========== 请求结构 ==========

// VoteRequest 投票请求
type VoteRequest struct {
	ItemID string  `json:"item_id" yaml:"item_id"` // 项目ID
	Vote   *string `json:"vote" yaml:"vote"`       // 投票类型（"up" | "down" | null）
}

// CommentRequest 评论请求
type CommentRequest struct {
	ItemID  string `json:"item_id" yaml:"item_id"` // 项目ID
	Content string `json:"content" yaml:"content"` // 评论内容
}

// SubmitSuggestionRequest 提交新建议请求
type SubmitSuggestionRequest struct {
	Title       string   `json:"title" yaml:"title"`             // 标题
	Description string   `json:"description" yaml:"description"` // 描述
	Priority    Priority `json:"priority" yaml:"priority"`       // 优先级
}

// ========== 服务实现 ==========

// RoadmapService 路线图服务
type RoadmapService struct {
	app        *application.App
	configFile string
	config     *RoadmapConfig // 内存中的配置缓存
	mu         sync.RWMutex   // 保护 config 和文件操作
}

// NewRoadmapService 创建路线图服务实例
func NewRoadmapService(app *application.App) *RoadmapService {
	s := &RoadmapService{
		app:        app,
		configFile: getConfigFilePath(),
	}

	// 尝试加载配置，如果不存在则创建默认配置
	config, err := s.loadConfig()
	if err != nil {
		config = s.getDefaultConfig()
		if saveErr := s.saveConfig(config); saveErr != nil {
			// 保存失败时记录日志，但不阻止服务启动
			fmt.Fprintf(os.Stderr, "[roadmap] 保存默认配置失败: %v\n", saveErr)
		}
	}
	s.config = config

	return s
}

// getConfigFilePath 获取配置文件路径
// 优先使用用户配置目录，确保开发和生产环境一致性
func getConfigFilePath() string {
	// 尝试获取用户配置目录
	configDir, err := os.UserConfigDir()
	if err != nil {
		// 降级到可执行文件目录
		return getExecDirConfigPath()
	}

	// 使用用户配置目录下的应用专用目录
	appConfigDir := filepath.Join(configDir, "rain-qa-func")
	return filepath.Join(appConfigDir, ConfigFileName)
}

// getExecDirConfigPath 获取可执行文件目录下的配置路径（降级方案）
func getExecDirConfigPath() string {
	execPath, err := os.Executable()
	if err != nil {
		return filepath.Join(".", ConfigFileName)
	}
	return filepath.Join(filepath.Dir(execPath), ConfigFileName)
}

// getDefaultConfig 获取默认配置
func (s *RoadmapService) getDefaultConfig() *RoadmapConfig {
	now := time.Now().Unix()

	return &RoadmapConfig{
		Version: "1.0",
		Items: []RoadmapItem{
			{
				ID:          generateID(),
				Title:       "战斗测试用例可视化编辑器",
				Description: "提供拖拽式界面，方便QA配置复杂的战斗测试场景，支持技能选择、属性配置、战力设置等。",
				Status:      StatusPlanning,
				Priority:    PriorityHigh,
				Author:      getSystemUser(),
				CreatedAt:   now,
				UpdatedAt:   now,
				Votes:       Votes{Up: 5, Down: 0},
				Comments:    []Comment{},
			},
			{
				ID:          generateID(),
				Title:       "配表检查规则可视化配置",
				Description: "通过界面配置配表检查规则，无需修改代码，支持自定义校验逻辑。",
				Status:      StatusInProgress,
				Priority:    PriorityMedium,
				Author:      getSystemUser(),
				CreatedAt:   now - 86400,
				UpdatedAt:   now - 86400,
				Votes:       Votes{Up: 3, Down: 1},
				Comments: []Comment{
					{
						ID:        generateID(),
						Author:    getSystemUser(),
						Content:   "这个功能很实用！",
						CreatedAt: now - 3600,
					},
				},
			},
			{
				ID:          generateID(),
				Title:       "测试报告自动生成",
				Description: "自动生成美观的测试报告，支持导出PDF、Excel格式，包含图表和统计数据。",
				Status:      StatusPlanning,
				Priority:    PriorityLow,
				Author:      getSystemUser(),
				CreatedAt:   now - 172800,
				UpdatedAt:   now - 172800,
				Votes:       Votes{Up: 8, Down: 0},
				Comments:    []Comment{},
			},
		},
	}
}

// generateID 生成唯一ID（使用加密安全的随机数）
func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// 降级到时间戳+随机数
		return fmt.Sprintf("id_%d", time.Now().UnixNano())
	}
	return "id_" + hex.EncodeToString(b)
}

// getSystemUser 获取系统用户名
func getSystemUser() string {
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	if u := os.Getenv("USERNAME"); u != "" {
		return u
	}
	return "系统"
}

// loadConfig 加载配置文件
func (s *RoadmapService) loadConfig() (*RoadmapConfig, error) {
	data, err := os.ReadFile(s.configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config RoadmapConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// saveConfig 保存配置文件到磁盘
func (s *RoadmapService) saveConfig(config *RoadmapConfig) error {
	// 确保目录存在
	dir := filepath.Dir(s.configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(s.configFile, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// persist 将内存中的配置持久化到磁盘
func (s *RoadmapService) persist() error {
	return s.saveConfig(s.config)
}

// GetItems 获取所有路线图项目
// @frontend
func (s *RoadmapService) GetItems() ([]RoadmapItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回深拷贝，防止外部修改影响内存缓存
	result := make([]RoadmapItem, len(s.config.Items))
	for i, item := range s.config.Items {
		result[i] = item
		result[i].Comments = make([]Comment, len(item.Comments))
		copy(result[i].Comments, item.Comments)
	}
	return result, nil
}

// GetItem 获取单个路线图项目
// @frontend
func (s *RoadmapService) GetItem(id string) (*RoadmapItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.config.Items {
		if s.config.Items[i].ID == id {
			// 返回副本
			item := s.config.Items[i]
			return &item, nil
		}
	}

	return nil, fmt.Errorf("未找到项目: %s", id)
}

// Vote 投票
// @frontend
func (s *RoadmapService) Vote(req VoteRequest) (*RoadmapItem, error) {
	if req.ItemID == "" {
		return nil, fmt.Errorf("项目ID不能为空")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找项目
	var item *RoadmapItem
	for i := range s.config.Items {
		if s.config.Items[i].ID == req.ItemID {
			item = &s.config.Items[i]
			break
		}
	}

	if item == nil {
		return nil, fmt.Errorf("未找到项目: %s", req.ItemID)
	}

	// 处理投票
	oldVote := item.Votes.UserVote

	// 如果之前有投票，先移除
	if oldVote != nil {
		switch *oldVote {
		case "up":
			item.Votes.Up--
		case "down":
			item.Votes.Down--
		}
	}

	// 如果新投票不为空，添加新投票
	if req.Vote != nil && *req.Vote != "" {
		switch *req.Vote {
		case "up":
			item.Votes.Up++
		case "down":
			item.Votes.Down++
		default:
			return nil, fmt.Errorf("无效的投票类型: %s", *req.Vote)
		}
		item.Votes.UserVote = req.Vote
	} else {
		item.Votes.UserVote = nil
	}

	item.UpdatedAt = time.Now().Unix()

	// 持久化到磁盘
	if err := s.persist(); err != nil {
		return nil, err
	}

	// 返回副本
	result := *item
	return &result, nil
}

// AddComment 添加评论
// @frontend
func (s *RoadmapService) AddComment(req CommentRequest) (*RoadmapItem, error) {
	if req.ItemID == "" {
		return nil, fmt.Errorf("项目ID不能为空")
	}
	if req.Content == "" {
		return nil, fmt.Errorf("评论内容不能为空")
	}
	if len(req.Content) > MaxCommentLength {
		return nil, fmt.Errorf("评论内容不能超过%d个字符", MaxCommentLength)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找项目
	var item *RoadmapItem
	for i := range s.config.Items {
		if s.config.Items[i].ID == req.ItemID {
			item = &s.config.Items[i]
			break
		}
	}

	if item == nil {
		return nil, fmt.Errorf("未找到项目: %s", req.ItemID)
	}

	// 创建新评论
	comment := Comment{
		ID:        generateID(),
		Author:    getSystemUser(),
		Content:   req.Content,
		CreatedAt: time.Now().Unix(),
	}

	item.Comments = append(item.Comments, comment)
	item.UpdatedAt = time.Now().Unix()

	// 持久化到磁盘
	if err := s.persist(); err != nil {
		return nil, err
	}

	// 返回副本
	result := *item
	return &result, nil
}

// isValidPriority 检查优先级是否有效
func isValidPriority(p Priority) bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	}
	return false
}

// isValidStatus 检查状态是否有效
func isValidStatus(s RoadmapStatus) bool {
	switch s {
	case StatusPlanning, StatusInProgress, StatusCompleted, StatusRejected:
		return true
	}
	return false
}

// SubmitSuggestion 提交新建议
// @frontend
func (s *RoadmapService) SubmitSuggestion(req SubmitSuggestionRequest) (*RoadmapItem, error) {
	if req.Title == "" || len(req.Title) > MaxTitleLength {
		return nil, fmt.Errorf("标题不能为空且不能超过%d个字符", MaxTitleLength)
	}
	if req.Description == "" || len(req.Description) > MaxDescriptionLength {
		return nil, fmt.Errorf("描述不能为空且不能超过%d个字符", MaxDescriptionLength)
	}
	if !isValidPriority(req.Priority) {
		return nil, fmt.Errorf("无效的优先级: %s", req.Priority)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()

	// 创建新项目
	newItem := RoadmapItem{
		ID:          generateID(),
		Title:       req.Title,
		Description: req.Description,
		Status:      StatusPlanning,
		Priority:    req.Priority,
		Author:      getSystemUser(),
		CreatedAt:   now,
		UpdatedAt:   now,
		Votes:       Votes{Up: 0, Down: 0},
		Comments:    make([]Comment, 0),
	}

	// 添加到列表开头
	s.config.Items = append([]RoadmapItem{newItem}, s.config.Items...)

	// 持久化到磁盘
	if err := s.persist(); err != nil {
		return nil, err
	}

	return &newItem, nil
}

// UpdateStatus 更新项目状态
// @frontend
func (s *RoadmapService) UpdateStatus(id string, status RoadmapStatus) (*RoadmapItem, error) {
	if id == "" {
		return nil, fmt.Errorf("项目ID不能为空")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找项目
	var item *RoadmapItem
	for i := range s.config.Items {
		if s.config.Items[i].ID == id {
			item = &s.config.Items[i]
			break
		}
	}

	if item == nil {
		return nil, fmt.Errorf("未找到项目: %s", id)
	}

	if !isValidStatus(status) {
		return nil, fmt.Errorf("无效的状态: %s", status)
	}

	item.Status = status
	item.UpdatedAt = time.Now().Unix()

	// 持久化到磁盘
	if err := s.persist(); err != nil {
		return nil, err
	}

	// 返回副本
	result := *item
	return &result, nil
}

// ReloadConfig 重新加载配置文件
// @frontend
func (s *RoadmapService) ReloadConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, err := s.loadConfig()
	if err != nil {
		return err
	}

	s.config = config
	return nil
}

// GetConfigFilePath 获取配置文件路径
// @frontend
func (s *RoadmapService) GetConfigFilePath() string {
	return s.configFile
}

// ExportToJSON 导出为JSON格式
// @frontend
func (s *RoadmapService) ExportToJSON() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("导出JSON失败: %w", err)
	}

	return string(data), nil
}
