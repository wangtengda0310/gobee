package chat

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
	"github.com/wangtengda0310/gobee/agent/pkg/llm/anthropic"
	"github.com/wangtengda0310/gobee/agent/pkg/tool"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
)

//go:embed all:chatui
var chatUI embed.FS

// buildSystemPrompt 构建包含平台信息的 system prompt
func buildSystemPrompt() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	// 根据平台提供命令示例
	var cmdExamples string
	switch osName {
	case "windows":
		cmdExamples = `当前是 Windows 系统，请使用 Windows 命令：
- 列出目录: dir 或 dir D:\
- 查看文件: type filename
- 文件信息: dir /s
- 环境变量: echo %PATH%
- 进程列表: tasklist`
	default:
		cmdExamples = `当前是 Linux/macOS 系统，请使用 Unix 命令：
- 列出目录: ls -la
- 查看文件: cat filename
- 文件查找: find /path -name "*.txt"
- 磁盘使用: df -h
- 进程列表: ps aux`
	}

	return fmt.Sprintf(`你是 Exporter 工具执行助手。你可以帮助用户通过自然语言执行各种命令行工具。

## 运行环境
- 操作系统: %s
- 架构: %s

%s

## 可用工具
- execute_command: 执行 exporter 注册的命令
- list_commands: 列出所有可用命令
- get_task_result: 获取命令执行结果
- cancel_task: 取消正在执行的任务
- run_shell: 执行本地 shell 命令并返回输出结果

## 注意事项
- 请根据当前操作系统选择合适的命令语法
- run_shell 执行失败时（exitCode != 0），请检查 stderr 中的错误信息并调整命令
- 不要重复尝试相同的失败命令

请用中文回复用户。当用户请求执行操作时，选择合适的工具完成。`, osName, arch, cmdExamples)
}

const maxToolRounds = 10

// ChatHandler 聊天 handler 集合
type ChatHandler struct {
	config       *ClaudeConfig
	sessions     *SessionManager
	tools        *tool.Registry
	client       llm.Client
	currentModel string
}

// NewChatHandler 创建聊天 handler
func NewChatHandler(cfg *ClaudeConfig) (*ChatHandler, error) {
	client, err := anthropic.NewClient(
		anthropic.WithAPIKey(cfg.APIKey),
		anthropic.WithBaseURL(cfg.BaseURL),
		anthropic.WithModel(cfg.DefaultModel),
		anthropic.WithMaxTokens(4096),
		anthropic.WithTimeout(300*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 LLM 客户端失败: %w", err)
	}

	return &ChatHandler{
		config:       cfg,
		sessions:     NewSessionManager(),
		tools:        NewExporterTools(),
		client:       client,
		currentModel: cfg.DefaultModel,
	}, nil
}

// RegisterHandlers 注册所有 /chat/ 路由到 mux
func (h *ChatHandler) RegisterHandlers(mux *http.ServeMux) {
	// 前端页面
	registerRoute(RouteInfo{Path: "/chat/", Method: "GET", Summary: "聊天页面"})
	mux.HandleFunc("/chat/", h.handleChatPage)

	// 帮助页面
	registerRoute(RouteInfo{Path: "/chat/help", Method: "GET", Summary: "API 帮助页面"})
	mux.HandleFunc("/chat/help", h.handleHelp)

	// API 端点
	registerRoute(RouteInfo{Path: "/chat/api/message", Method: "POST", Summary: "发送消息（SSE流式响应）"})
	mux.HandleFunc("/chat/api/message", h.handleMessage)

	registerRoute(RouteInfo{Path: "/chat/api/config", Method: "GET", Summary: "获取模型配置"})
	registerRoute(RouteInfo{Path: "/chat/api/config", Method: "PUT", Summary: "切换模型"})
	mux.HandleFunc("/chat/api/config", h.handleConfig)

	registerRoute(RouteInfo{Path: "/chat/api/history", Method: "GET", Summary: "获取会话历史"})
	registerRoute(RouteInfo{Path: "/chat/api/history", Method: "DELETE", Summary: "清空会话历史"})
	mux.HandleFunc("/chat/api/history", h.handleHistory)
}

// handleChatPage 处理前端页面请求
func (h *ChatHandler) handleChatPage(w http.ResponseWriter, r *http.Request) {
	// 根据请求路径确定文件名
	path := r.URL.Path
	switch {
	case path == "/chat/" || path == "/chat":
		path = "chatui/index.html"
	case strings.HasPrefix(path, "/chat/"):
		path = "chatui/" + strings.TrimPrefix(path, "/chat/")
	}

	// 从嵌入的文件系统读取文件
	data, err := chatUI.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 设置 Content-Type
	switch {
	case strings.HasSuffix(path, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	}

	w.Write(data)
}

// handleHelp 处理帮助页面请求
func (h *ChatHandler) handleHelp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	routes := GetRoutes()

	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Exporter Chat API 帮助</title>
    <style>
        body { font-family: system-ui, sans-serif; max-width: 900px; margin: 40px auto; padding: 0 20px; background: #1a1a2e; color: #e0e0e0; }
        h1 { color: #e94560; border-bottom: 2px solid #0f3460; padding-bottom: 10px; }
        h2 { color: #e94560; margin-top: 30px; }
        .endpoint { background: #16213e; padding: 15px; margin: 15px 0; border-radius: 8px; border-left: 4px solid #e94560; }
        .method { display: inline-block; padding: 4px 8px; border-radius: 4px; font-weight: bold; margin-right: 10px; }
        .get { background: #4CAF50; color: white; }
        .post { background: #2196F3; color: white; }
        .put { background: #FF9800; color: white; }
        .delete { background: #f44336; color: white; }
        .path { font-family: monospace; font-size: 14px; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { padding: 8px; text-align: left; border-bottom: 1px solid #0f3460; }
        th { background: #0f3460; }
        code { background: #0d1b2a; padding: 2px 6px; border-radius: 3px; font-family: monospace; }
        pre { background: #0d1b2a; padding: 15px; border-radius: 6px; overflow-x: auto; }
        pre code { background: none; padding: 0; }
    </style>
</head>
<body>
    <h1>Exporter Chat API 帮助</h1>
    <p>Exporter Chat Agent 提供了 RESTful API 和 SSE 流式聊天接口。</p>
`

	for _, route := range routes {
		methodClass := strings.ToLower(route.Method)
		html += fmt.Sprintf(`
    <div class="endpoint">
        <span class="method %s">%s</span>
        <span class="path">%s</span>
        <p>%s</p>
`, methodClass, route.Method, route.Path, route.Summary)

		if route.Description != "" {
			html += fmt.Sprintf("        <p><small>%s</small></p>\n", route.Description)
		}

		if len(route.Params) > 0 {
			html += `        <table>
            <thead><tr><th>参数名</th><th>类型</th><th>必填</th><th>说明</th></tr></thead>
            <tbody>
`
			for _, param := range route.Params {
				required := "否"
				if param.Required {
					required = "是"
				}
				html += fmt.Sprintf("                <tr><td><code>%s</code></td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
					param.Name, param.Type, required, param.Desc)
			}
			html += `            </tbody>
        </table>
`
		}

		if route.Response != "" {
			html += fmt.Sprintf("        <p><strong>响应格式:</strong> <code>%s</code></p>\n", route.Response)
		}

		html += "    </div>\n"
	}

	html += `
    <h2>使用示例</h2>
    <pre><code>// 发送消息（SSE 流式）
curl -N -X POST http://localhost:8080/chat/api/message \
  -H "Content-Type: application/json" \
  -d '{"content":"列出所有命令","session_id":"my-session"}'

// 切换模型
curl -X PUT http://localhost:8080/chat/api/config \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-3-5-sonnet-20241022"}'

// 清空历史
curl -X DELETE "http://localhost:8080/chat/api/history?session_id=my-session"
</code></pre>
</body>
</html>
`

	w.Write([]byte(html))
}

// handleConfig 处理配置相关请求
func (h *ChatHandler) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch r.Method {
	case http.MethodGet:
		// 获取配置
		models := make([]string, 0, len(h.config.Models))
		for _, m := range h.config.Models {
			if m != "" {
				models = append(models, m)
			}
		}

		json.NewEncoder(w).Encode(map[string]any{
			"current": h.currentModel,
			"models":  models,
		})

	case http.MethodPut:
		// 切换模型
		var req struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Model == "" {
			http.Error(w, "Model name is required", http.StatusBadRequest)
			return
		}

		// 创建新的客户端
		client, err := anthropic.NewClient(
			anthropic.WithAPIKey(h.config.APIKey),
			anthropic.WithBaseURL(h.config.BaseURL),
			anthropic.WithModel(req.Model),
			anthropic.WithMaxTokens(4096),
			anthropic.WithTimeout(300*time.Second),
		)
		if err != nil {
			logger.Error("创建新的 LLM 客户端失败: %v", err)
			http.Error(w, "Failed to switch model", http.StatusInternalServerError)
			return
		}

		h.currentModel = req.Model
		h.client = client

		logger.Info("已切换模型: %s", req.Model)
		json.NewEncoder(w).Encode(map[string]any{
			"current": h.currentModel,
			"status":  "switched",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleHistory 处理会话历史相关请求
func (h *ChatHandler) handleHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// 获取历史
		messages := h.sessions.GetMessages(sessionID)
		json.NewEncoder(w).Encode(map[string]any{
			"session_id": sessionID,
			"messages":   messages,
			"count":      len(messages),
		})

	case http.MethodDelete:
		// 清空历史
		h.sessions.ClearSession(sessionID)
		logger.Info("已清空会话: %s", sessionID)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "cleared",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMessage 处理聊天消息（SSE 流式）
// 流程：用户消息 → LLM 流式响应 → 若有工具调用则执行工具 → 再次调用 LLM → 循环直到无工具调用
func (h *ChatHandler) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 设置 SSE headers
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 解析请求
	var req struct {
		Content   string `json:"content"`
		Model     string `json:"model,omitempty"`
		SessionID string `json:"session_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSSE(w, "error", `{"message":"Invalid request body: `+jsonEscape(err.Error())+`"}`)
		return
	}

	if req.Content == "" {
		writeSSE(w, "error", `{"message":"Content is required"}`)
		return
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	// 如果指定了模型且与当前不同，切换模型
	if req.Model != "" && req.Model != h.currentModel {
		client, err := anthropic.NewClient(
			anthropic.WithAPIKey(h.config.APIKey),
			anthropic.WithBaseURL(h.config.BaseURL),
			anthropic.WithModel(req.Model),
			anthropic.WithMaxTokens(4096),
			anthropic.WithTimeout(300*time.Second),
		)
		if err != nil {
			logger.Error("切换模型失败: %v", err)
		} else {
			h.currentModel = req.Model
			h.client = client
		}
	}

	// 追加用户消息到会话
	userMsg := &llm.Message{
		Role:    llm.RoleUser,
		Content: llm.Text(req.Content),
	}
	h.sessions.AddMessage(sessionID, userMsg)

	// 获取历史消息
	history := h.sessions.GetMessages(sessionID)

	// 构建请求
	chatReq := &llm.ChatRequest{
		Messages: history,
		System:   buildSystemPrompt(),
		Tools:    h.tools.GetDefinitions(),
	}

	// 执行工具调用循环（最多 maxToolRounds 轮）
	ctx := r.Context()
	for round := 0; round < maxToolRounds; round++ {
		// 调用 LLM 流式接口
		chunkChan, err := h.client.Stream(ctx, chatReq)
		if err != nil {
			logger.Error("LLM 调用失败: %v", err)
			writeSSE(w, "error", `{"message":"LLM call failed: `+jsonEscape(err.Error())+`"}`)
			return
		}

		// 收集流式响应
		// anthropic client 的 processStream 已在内部累积 ToolCalls，
		// 完整数据在 DoneChunk.Response 中返回
		var fullContent strings.Builder
		var lastResponse *llm.ChatResponse

		for chunk := range chunkChan {
			if chunk.IsError() {
				logger.Error("流式处理错误: %v", chunk.Error)
				writeSSE(w, "error", `{"message":"Stream error: `+jsonEscape(chunk.Error.Error())+`"}`)
				return
			}

			switch chunk.Type {
			case llm.ChunkTypeContent:
				fullContent.WriteString(chunk.Content)
				writeSSE(w, "content", fmt.Sprintf(`{"text":"%s"}`, jsonEscape(chunk.Content)))
			case llm.ChunkTypeDone:
				lastResponse = chunk.Response
			}
		}

		// 检查是否需要工具调用（使用 lastResponse 中的完整 ToolCalls）
		if lastResponse != nil && lastResponse.StopReason == llm.StopReasonToolUse && len(lastResponse.ToolCalls) > 0 {
			// 先添加 assistant 消息到会话和请求（包含工具调用信息）
			// 这样下一轮 LLM 才能看到完整的 assistant + tool_call 上下文
			assistantMsg := &llm.Message{
				Role:      llm.RoleAssistant,
				Content:   llm.Text(fullContent.String()),
				ToolCalls: lastResponse.ToolCalls,
			}
			h.sessions.AddMessage(sessionID, assistantMsg)
			chatReq.Messages = append(chatReq.Messages, assistantMsg)

			// 执行每个工具调用
			for _, tc := range lastResponse.ToolCalls {
				writeSSE(w, "tool_use", fmt.Sprintf(`{"name":"%s","args":%s}`,
					tc.Function.Name, tc.Function.Arguments))

				// 解析参数 JSON 字符串为 map
				var args map[string]any
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					logger.Error("解析工具参数失败: %v", err)
					writeSSE(w, "tool_result", fmt.Sprintf(`{"name":"%s","success":false,"result":"%s"}`,
						tc.Function.Name, jsonEscape(err.Error())))
					continue
				}

				// 执行工具
				result, err := h.tools.Execute(ctx, tc.Function.Name, args)
				if err != nil {
					logger.Error("工具执行失败: %v", err)
					writeSSE(w, "tool_result", fmt.Sprintf(`{"name":"%s","success":false,"result":"%s"}`,
						tc.Function.Name, jsonEscape(err.Error())))
					continue
				}

				// 格式化结果并发送 SSE
				resultJSON, _ := json.Marshal(result)
				writeSSE(w, "tool_result", fmt.Sprintf(`{"name":"%s","success":true,"result":%s}`,
					tc.Function.Name, string(resultJSON)))

				// 添加工具结果到请求消息（下一轮 LLM 需要）
				toolResultMsg := &llm.Message{
					Role:       llm.RoleTool,
					ToolCallID: tc.ID,
					Name:       tc.Function.Name,
					Content:    llm.Text(string(resultJSON)),
				}
				chatReq.Messages = append(chatReq.Messages, toolResultMsg)
			}

			// 继续下一轮 LLM 调用
			continue
		}

		// 正常结束（无工具调用），添加助手消息到会话
		if fullContent.Len() > 0 {
			assistantMsg := &llm.Message{
				Role:    llm.RoleAssistant,
				Content: llm.Text(fullContent.String()),
			}
			h.sessions.AddMessage(sessionID, assistantMsg)
		}

		writeSSE(w, "done", `{}`)
		return
	}

	// 达到最大轮数
	writeSSE(w, "error", `{"message":"Maximum tool rounds exceeded"}`)
}

// writeSSE 发送 SSE 事件
func writeSSE(w http.ResponseWriter, event, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// jsonEscape 转义 JSON 字符串
func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}

// generateSessionID 生成会话 ID
func generateSessionID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
