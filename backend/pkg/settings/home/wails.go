package home

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	exceltest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/excel-test"
	prototest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"
	"github.com/wangtengda0310/gobee/agent/pkg/agent"
	"github.com/wangtengda0310/gobee/agent/pkg/llm"
	"github.com/wangtengda0310/gobee/agent/pkg/memory"
	"github.com/wangtengda0310/gobee/agent/pkg/tool"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ChatService 聊天服务
type ChatService struct {
	app           *application.App // Wails 应用实例，仅构造时使用
	emitter       EventEmitter     // 事件发射器，所有 Event.Emit 调用走这里
	configFile    string
	config        *ChatConfig
	memory        *memoryAdapter
	registry      *tool.Registry    // 工具注册表（用于获取工具定义）
	currentClient llm.ChatCompleter // 当前 LLM 客户端
	agent         *agent.Agent      // Agent 实例
	cancelFunc    context.CancelFunc
	mu            sync.RWMutex
}

// NewChatService 创建聊天服务
func NewChatService(app *application.App) *ChatService {
	return NewChatServiceWithServices(app, nil)
}

// NewChatServiceWithServices 创建聊天服务（带依赖注入）
func NewChatServiceWithServices(app *application.App, services *ChatServices) *ChatService {
	// 获取可执行文件目录作为配置文件存储位置
	execPath, _ := os.Executable()
	baseDir := filepath.Dir(execPath)

	// 使用官方 FileMemory，并尝试加载旧格式
	filePath := filepath.Join(baseDir, "chat_history.json")
	mem := memory.NewFileMemory(100, filePath, memory.WithPreserveSystem(true))

	// 尝试加载文件（如果存在）
	ctx := context.Background()
	if err := mem.Load(ctx); err != nil {
		log.Printf("[ChatService] 加载历史记录: %v", err)
	}

	// 初始化事件发射器：app.Event 是 *application.EventManager，隐式实现 EventEmitter 接口
	var emitter EventEmitter = &noopEventEmitter{}
	if app != nil && app.Event != nil {
		emitter = app.Event
	}

	s := &ChatService{
		app:        app,
		emitter:    emitter,
		configFile: filepath.Join(baseDir, "app_chat_config.json"),
		memory: &memoryAdapter{
			FileMemory: mem,
		},
		registry: tool.NewRegistry(),
	}

	// 注册业务模块 LLM 工具
	if services != nil && services.ExcelCheckService != nil {
		exceltest.InitLLMTools(s.registry, services.ExcelCheckService)
	}
	if services != nil && services.RecordControlService != nil && services.TestCaseService != nil && services.RecordFileService != nil {
		prototest.InitLLMTools(s.registry, services.RecordControlService, services.TestCaseService, services.RecordFileService)
	}

	return s
}

// GetConfig 获取当前配置
// @frontend
func (s *ChatService) GetConfig() (*ChatConfig, error) {
	s.mu.RLock()
	if s.config != nil {
		defer s.mu.RUnlock()
		return s.config, nil
	}
	s.mu.RUnlock()

	// 尝试从文件加载配置
	config, err := s.loadConfig()
	if err != nil {
		return s.getDefaultConfig(), nil
	}

	s.mu.Lock()
	s.config = config
	s.mu.Unlock()

	return config, nil
}

// SaveConfig 保存配置
// @frontend
func (s *ChatService) SaveConfig(config *ChatConfig) error {
	s.mu.Lock()
	s.config = config
	s.currentClient = nil
	s.agent = nil
	s.mu.Unlock()

	return s.saveConfig(config)
}

// SendMessageStream 流式发送消息
// @frontend
func (s *ChatService) SendMessageStream(content string) error {
	log.Printf("[ChatService] SendMessageStream 开始, 内容长度: %d", len(content))

	// 先停止之前的请求
	s.StopStream()

	s.mu.Lock()
	defer s.mu.Unlock()

	// 确保配置存在
	if s.config == nil {
		var err error
		s.config, err = s.loadConfig()
		if err != nil {
			s.config = s.getDefaultConfig()
		}
	}

	log.Printf("[ChatService] 当前提供商: %s", s.config.Provider)

	// 获取或创建 Agent
	ag, err := s.getOrCreateAgent()
	if err != nil {
		log.Printf("[ChatService] 创建 Agent 失败: %v", err)
		s.emitter.Emit("chatStreamError", err.Error())
		return err
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	s.cancelFunc = cancel

	// 使用 Agent.RunStream 执行任务
	log.Printf("[ChatService] 启动 Agent 流式处理")
	eventCh, err := ag.RunStream(ctx, content)
	if err != nil {
		log.Printf("[ChatService] Agent 启动失败: %v", err)
		s.emitter.Emit("chatStreamError", err.Error())
		return err
	}

	// 启动 goroutine 处理 Agent 事件
	go s.processAgentEvents(ctx, eventCh)

	return nil
}

// StopStream 停止流式生成
// @frontend
func (s *ChatService) StopStream() {
	s.mu.Lock()
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}
	s.mu.Unlock()
}

// GetHistory 获取历史记录
// @frontend
func (s *ChatService) GetHistory() (*ChatSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ctx := context.Background()
	if s.memory.Len() == 0 {
		if err := s.memory.Load(ctx); err != nil {
			return nil, err
		}
	}

	messages := s.memory.GetMessages()
	return &ChatSession{
		ID:        fmt.Sprintf("session_%d", time.Now().Unix()),
		Messages:  messages,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}, nil
}

// ClearHistory 清空历史
// @frontend
func (s *ChatService) ClearHistory() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	return s.memory.Clear(ctx)
}

// ImportFromClaudeCode 从 Claude Code 配置导入
// @frontend
func (s *ChatService) ImportFromClaudeCode() (*ChatConfig, error) {
	claudeSettings, err := settings.LoadClaudeCodeConfig()
	if err != nil {
		return nil, fmt.Errorf("加载 Claude Code 配置失败: %w", err)
	}

	cfg, err := s.loadConfig()
	if err != nil {
		cfg = s.getDefaultConfig()
	}

	applyClaudeCodeConfig(claudeSettings, cfg)

	s.mu.Lock()
	s.config = cfg
	s.currentClient = nil
	s.agent = nil
	s.mu.Unlock()

	return cfg, nil
}
