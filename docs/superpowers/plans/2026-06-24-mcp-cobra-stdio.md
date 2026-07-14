# MCP stdio 子命令实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `backend/pkg/settings/mcp/cobra.go` 实现为 stdio 传输的 MCP 服务器入口，复用与 HTTP 版本完全一致的业务工具集。

**Architecture:** 提取 `services.go` 统一组装 MCP 服务（分 `StdioMcpServices` 和 `HttpMcpServices`），HTTP 服务器和 stdio 子命令分别复用；stdio 服务器通过 `modelcontextprotocol/go-sdk` 的 stdio 支持直接处理 stdin/stdout 上的 JSON-RPC。

**Tech Stack:** Go 1.25, Cobra, modelcontextprotocol/go-sdk, Wails v3 application.Service

## Global Constraints

- 代码使用中文注释，对外暴露方法需要中文注释。
- 修改代码需同步修改注释。
- 超过 20 行的方法需补充以流程为视角的注释。
- Go 单元测试使用 testify/assert。
- 禁止编写欺骗性单元测试。
- 每次代码修改后需同步相关文档。
- 单个源文件中超过 5 个方法考虑拆分。
- 使用绝对路径修改文件。
- 任务完成后运行 IDE 诊断和构建命令验证。
- 提交前运行 `wails3 task build` 构建可执行程序。
- 相关 CLAUDE.md 必须同步更新。

---

## File Map

| 文件 | 职责 |
|---|---|
| `backend/pkg/settings/mcp/services.go` | 新增：定义 `StdioMcpServices` 和 `HttpMcpServices`，提供 `BuildDefaultStdioMcpServices()` 和 `BuildDefaultHttpMcpServices()` |
| `backend/pkg/settings/mcp/server.go` | 修改：`MCPServer` 改用 `*HttpMcpServices`，`registerTools()` 按 `HttpMcpServices` 字段注册工具 |
| `backend/pkg/settings/mcp/cobra.go` | 修改：`NewMCPCmd()` 启动 stdio MCP 服务器 |
| `backend/pkg/settings/mcp/cobra-help.md` | 修改：更新帮助文档，说明 stdio 用途和示例 |
| `cmd/rain-qa-func/wails.go` | 修改：用 `BuildDefaultHttpMcpServices()` 替换 inline 服务组装 |
| `backend/pkg/settings/mcp/services_test.go` | 新增：测试两个构建函数返回非 nil 且字段完整 |
| `backend/pkg/settings/mcp/cobra_test.go` | 新增：测试 stdio 服务器创建和 tools/list 非空 |
| `backend/pkg/settings/mcp/CLAUDE.md` | 更新：补充服务容器、stdio 子命令说明 |

---

### Task 1: 提取服务容器到 services.go

**Files:**
- Create: `backend/pkg/settings/mcp/services.go`
- Modify: `backend/pkg/settings/mcp/server.go:27-42`（原 `Services` 结构体删除）
- Test: `backend/pkg/settings/mcp/services_test.go`

**Interfaces:**
- Consumes: 无
- Produces:
  - `type StdioMcpServices struct { ... }`
  - `type HttpMcpServices struct { ... }`
  - `func BuildDefaultStdioMcpServices() (*StdioMcpServices, error)`
  - `func BuildDefaultHttpMcpServices() (*HttpMcpServices, error)`

- [ ] **Step 1: 编写 services.go**

创建 `backend/pkg/settings/mcp/services.go`，内容如下：

```go
package mcp

import (
	activitywikicheck "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game"
	mcputil "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/mcp"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/robotext"
	exceltest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/excel-test"
	functiontest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test"
	herowikicheck "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/hero-wiki-check"
	resourcechecker "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"
)

// StdioMcpServices stdio 模式下暴露的 MCP 服务集合
type StdioMcpServices struct {
	StdioMcpJsonCaseService          *functiontest.JsonCaseService
	StdioMcpFuncCaseConfigService    *functiontest.FuncCaseConfigService
	StdioMcpExcelCheckService        *exceltest.ExcelCheckService
	StdioMcpExcelConfigService       *exceltest.ExcelConfigService
	StdioMcpGameExcelService         *game.GameExcelService
	StdioMcpExcelTestGameService     *exceltest.ExcelTestGameService
	StdioMcpHeroResCheckService      *resourcechecker.HeroResCheckService
	StdioMcpHeroWikiResCheckService  *herowikicheck.HeroWikiResCheckService
	StdioMcpActivityWikiCheckService *activitywikicheck.ActivityWikiCheckService
	StdioMcpConfigService            *settings.MCPConfigService
	StdioMcpFeishuNotifyConfigService *settings.FeishuNotifyConfigService
	StdioMcpRobotExtService          *robotext.RobotExtService
}

// HttpMcpServices HTTP 模式下暴露的 MCP 服务集合
type HttpMcpServices struct {
	HttpMcpJsonCaseService          *functiontest.JsonCaseService
	HttpMcpFuncCaseConfigService    *functiontest.FuncCaseConfigService
	HttpMcpExcelCheckService        *exceltest.ExcelCheckService
	HttpMcpExcelConfigService       *exceltest.ExcelConfigService
	HttpMcpGameExcelService         *game.GameExcelService
	HttpMcpExcelTestGameService     *exceltest.ExcelTestGameService
	HttpMcpHeroResCheckService      *resourcechecker.HeroResCheckService
	HttpMcpHeroWikiResCheckService  *herowikicheck.HeroWikiResCheckService
	HttpMcpActivityWikiCheckService *activitywikicheck.ActivityWikiCheckService
	HttpMcpConfigService            *settings.MCPConfigService
	HttpMcpFeishuNotifyConfigService *settings.FeishuNotifyConfigService
	HttpMcpRobotExtService          *robotext.RobotExtService
}

// buildMcpServiceDeps 创建所有 MCP 服务实际依赖的私有 helper。
// 两个传输模式使用相同的底层 Service 实例创建逻辑，避免重复代码。
func buildMcpServiceDeps() (
	funcConfigService *functiontest.FuncCaseConfigService,
	jsonCaseService *functiontest.JsonCaseService,
	excelCheckService *exceltest.ExcelCheckService,
	excelConfigService *exceltest.ExcelConfigService,
	gameExcelService *game.GameExcelService,
	excelTestGameService *exceltest.ExcelTestGameService,
	heroResCheckService *resourcechecker.HeroResCheckService,
	heroWikiResCheckService *herowikicheck.HeroWikiResCheckService,
	activityWikiCheckService *activitywikicheck.ActivityWikiCheckService,
	mcpConfigService *settings.MCPConfigService,
	feishuNotifyConfigService *settings.FeishuNotifyConfigService,
	robotExtService *robotext.RobotExtService,
	err error,
) {
	funcConfigService = functiontest.NewFuncCaseConfigService()
	jsonCaseService = functiontest.NewJsonCaseService(nil)
	excelCheckService = exceltest.NewExcelCheckServiceWithConfig(funcConfigService)
	excelConfigService = exceltest.NewExcelConfigService()
	gameExcelService = game.NewGameExcelService()
	if funcConfig, cfgErr := funcConfigService.GetConfig(); cfgErr == nil && funcConfig != nil && funcConfig.ExcelResourcesDir != "" {
		_ = gameExcelService.InitExcel(funcConfig.ExcelResourcesDir)
	}
	excelTestGameService = exceltest.NewExcelTestGameService()
	heroResCheckService = resourcechecker.NewHeroResCheckService()
	heroWikiResCheckService = herowikicheck.NewHeroWikiResCheckService()
	activityWikiCheckService = activitywikicheck.NewActivityWikiCheckService()
	mcpConfigService = settings.NewMCPConfigService()
	feishuNotifyConfigService = settings.NewFeishuNotifyConfigService()
	robotExtService = robotext.NewRobotExtService()
	return
}

// BuildDefaultStdioMcpServices 创建 stdio 模式所需的默认 MCP 服务集合
func BuildDefaultStdioMcpServices() (*StdioMcpServices, error) {
	funcConfigService, jsonCaseService, excelCheckService, excelConfigService,
		gameExcelService, excelTestGameService, heroResCheckService, heroWikiResCheckService,
		activityWikiCheckService, mcpConfigService, feishuNotifyConfigService, robotExtService, err := buildMcpServiceDeps()
	if err != nil {
		return nil, err
	}
	return &StdioMcpServices{
		StdioMcpJsonCaseService:           jsonCaseService,
		StdioMcpFuncCaseConfigService:     funcConfigService,
		StdioMcpExcelCheckService:         excelCheckService,
		StdioMcpExcelConfigService:        excelConfigService,
		StdioMcpGameExcelService:          gameExcelService,
		StdioMcpExcelTestGameService:      excelTestGameService,
		StdioMcpHeroResCheckService:       heroResCheckService,
		StdioMcpHeroWikiResCheckService:   heroWikiResCheckService,
		StdioMcpActivityWikiCheckService:  activityWikiCheckService,
		StdioMcpConfigService:             mcpConfigService,
		StdioMcpFeishuNotifyConfigService: feishuNotifyConfigService,
		StdioMcpRobotExtService:           robotExtService,
	}, nil
}

// BuildDefaultHttpMcpServices 创建 HTTP 模式所需的默认 MCP 服务集合
func BuildDefaultHttpMcpServices() (*HttpMcpServices, error) {
	funcConfigService, jsonCaseService, excelCheckService, excelConfigService,
		gameExcelService, excelTestGameService, heroResCheckService, heroWikiResCheckService,
		activityWikiCheckService, mcpConfigService, feishuNotifyConfigService, robotExtService, err := buildMcpServiceDeps()
	if err != nil {
		return nil, err
	}
	return &HttpMcpServices{
		HttpMcpJsonCaseService:           jsonCaseService,
		HttpMcpFuncCaseConfigService:     funcConfigService,
		HttpMcpExcelCheckService:         excelCheckService,
		HttpMcpExcelConfigService:        excelConfigService,
		HttpMcpGameExcelService:          gameExcelService,
		HttpMcpExcelTestGameService:      excelTestGameService,
		HttpMcpHeroResCheckService:       heroResCheckService,
		HttpMcpHeroWikiResCheckService:   heroWikiResCheckService,
		HttpMcpActivityWikiCheckService:  activityWikiCheckService,
		HttpMcpConfigService:             mcpConfigService,
		HttpMcpFeishuNotifyConfigService: feishuNotifyConfigService,
		HttpMcpRobotExtService:           robotExtService,
	}, nil
}
```

- [ ] **Step 2: 删除 server.go 中的旧 Services 结构体**

修改 `backend/pkg/settings/mcp/server.go`，删除第 26-42 行的 `Services` 结构体定义。保留其他代码不变，下一任务再修改 `MCPServer` 字段类型。

- [ ] **Step 3: 编写 services_test.go**

创建 `backend/pkg/settings/mcp/services_test.go`：

```go
package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildDefaultStdioMcpServices(t *testing.T) {
	services, err := BuildDefaultStdioMcpServices()
	require.NoError(t, err)
	require.NotNil(t, services)

	assert.NotNil(t, services.StdioMcpJsonCaseService)
	assert.NotNil(t, services.StdioMcpFuncCaseConfigService)
	assert.NotNil(t, services.StdioMcpExcelCheckService)
	assert.NotNil(t, services.StdioMcpExcelConfigService)
	assert.NotNil(t, services.StdioMcpGameExcelService)
	assert.NotNil(t, services.StdioMcpExcelTestGameService)
	assert.NotNil(t, services.StdioMcpHeroResCheckService)
	assert.NotNil(t, services.StdioMcpHeroWikiResCheckService)
	assert.NotNil(t, services.StdioMcpActivityWikiCheckService)
	assert.NotNil(t, services.StdioMcpConfigService)
	assert.NotNil(t, services.StdioMcpFeishuNotifyConfigService)
	assert.NotNil(t, services.StdioMcpRobotExtService)
}

func TestBuildDefaultHttpMcpServices(t *testing.T) {
	services, err := BuildDefaultHttpMcpServices()
	require.NoError(t, err)
	require.NotNil(t, services)

	assert.NotNil(t, services.HttpMcpJsonCaseService)
	assert.NotNil(t, services.HttpMcpFuncCaseConfigService)
	assert.NotNil(t, services.HttpMcpExcelCheckService)
	assert.NotNil(t, services.HttpMcpExcelConfigService)
	assert.NotNil(t, services.HttpMcpGameExcelService)
	assert.NotNil(t, services.HttpMcpExcelTestGameService)
	assert.NotNil(t, services.HttpMcpHeroResCheckService)
	assert.NotNil(t, services.HttpMcpHeroWikiResCheckService)
	assert.NotNil(t, services.HttpMcpActivityWikiCheckService)
	assert.NotNil(t, services.HttpMcpConfigService)
	assert.NotNil(t, services.HttpMcpFeishuNotifyConfigService)
	assert.NotNil(t, services.HttpMcpRobotExtService)
}
```

- [ ] **Step 4: 运行测试**

Run: `go test ./backend/pkg/settings/mcp/ -run TestBuildDefault -v`
Expected: PASS（测试阶段 `MCPServer` 还未改，只是结构体删除，应无编译错误）

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/settings/mcp/services.go backend/pkg/settings/mcp/services_test.go backend/pkg/settings/mcp/server.go
git commit -m "refactor(mcp): extract StdioMcpServices and HttpMcpServices into services.go

- Add services.go with BuildDefaultStdioMcpServices/BuildDefaultHttpMcpServices
- Remove old Services struct from server.go
- Add unit tests for both service containers

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: 修改 MCPServer 使用 HttpMcpServices

**Files:**
- Modify: `backend/pkg/settings/mcp/server.go:45-73`, `server.go:215-325`
- Test: `backend/pkg/settings/mcp/server_test.go`

**Interfaces:**
- Consumes: `HttpMcpServices`（Task 1）
- Produces: `func NewMCPServer(config *settings.MCPConfig, services *HttpMcpServices) *MCPServer`

- [ ] **Step 1: 修改 MCPServer 字段类型和构造函数**

将 `backend/pkg/settings/mcp/server.go` 中：

```go
type MCPServer struct {
	server       *mcpgo.Server
	config       *settings.MCPConfig
	services     *Services
	...
}

func NewMCPServer(config *settings.MCPConfig, services *Services) *MCPServer {
```

改为：

```go
type MCPServer struct {
	server       *mcpgo.Server
	config       *settings.MCPConfig
	services     *HttpMcpServices
	...
}

func NewMCPServer(config *settings.MCPConfig, services *HttpMcpServices) *MCPServer {
```

- [ ] **Step 2: 修改 registerTools 使用新字段名**

将 `registerTools()` 中所有 `s.services.Xxx` 替换为 `s.services.HttpMcpXxx`。例如：

- `s.services.JsonCaseService` → `s.services.HttpMcpJsonCaseService`
- `s.services.FuncCaseConfigService` → `s.services.HttpMcpFuncCaseConfigService`
- `s.services.ExcelCheckService` → `s.services.HttpMcpExcelCheckService`
- `s.services.ExcelConfigService` → `s.services.HttpMcpExcelConfigService`
- `s.services.HeroWikiResCheckService` → `s.services.HttpMcpHeroWikiResCheckService`
- `s.services.ActivityWikiCheckService` → `s.services.HttpMcpActivityWikiCheckService`
- `s.services.ExcelTestGameService` → `s.services.HttpMcpExcelTestGameService`
- `s.services.GameExcelService` → `s.services.HttpMcpGameExcelService`
- `s.services.RobotExtService` → `s.services.HttpMcpRobotExtService`
- `s.services.FeishuNotifyConfigService` → `s.services.HttpMcpFeishuNotifyConfigService`
- `s.services.MCPConfigService` → `s.services.HttpMcpConfigService`

- [ ] **Step 3: 修改 server_test.go**

将 `TestNewMCPServer` 中的 `services := &Services{}` 改为 `services := &HttpMcpServices{}`。

- [ ] **Step 4: 编译验证**

Run: `go build ./backend/pkg/settings/mcp/`
Expected: 编译通过

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/settings/mcp/server.go backend/pkg/settings/mcp/server_test.go
git commit -m "refactor(mcp): MCPServer uses HttpMcpServices

- Update MCPServer field type and constructor signature
- Update registerTools to use HttpMcp-prefixed fields
- Update server_test to use HttpMcpServices

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: 修改 wails.go 使用 BuildDefaultHttpMcpServices

**Files:**
- Modify: `cmd/rain-qa-func/wails.go:199-225`

**Interfaces:**
- Consumes: `BuildDefaultHttpMcpServices()`（Task 1）
- Produces: 无

- [ ] **Step 1: 替换 inline 服务组装**

找到 `cmd/rain-qa-func/wails.go` 中以下代码（约 199-225 行）：

```go
// 创建 MCP 配置服务（用于前端绑定）
mcpConfigService := settings.NewMCPConfigService()
app.RegisterService(application.NewService(mcpConfigService))

// ========== 启动 MCP 服务器 ==========
mcpConfig, err := mcp.LoadConfig()
...
mcpFuncConfigService := functiontest.NewFuncCaseConfigService()

mcpServer := mcp.NewMCPServer(mcpConfig, &mcp.Services{
    JsonCaseService:           functiontest.NewJsonCaseService(app),
    FuncCaseConfigService:     mcpFuncConfigService,
    ExcelCheckService:         exceltest.NewExcelCheckServiceWithConfig(mcpFuncConfigService),
    ExcelConfigService:        exceltest.NewExcelConfigService(),
    GameExcelService:          gameExcelSvc,
    ExcelTestGameService:      exceltest.NewExcelTestGameService(),
    HeroResCheckService:       resourcechecker.NewHeroResCheckService(),
    HeroWikiResCheckService:   herowikicheck.NewHeroWikiResCheckService(),
    MCPConfigService:          mcpConfigService,
    FeishuNotifyConfigService: settings.NewFeishuNotifyConfigService(),
    RobotExtService:           robotext.NewRobotExtService(),
})

mcpConfigService.SetServer(mcpServer)
```

替换为：

```go
// 创建 MCP 配置服务（用于前端绑定）
mcpConfigService := settings.NewMCPConfigService()
app.RegisterService(application.NewService(mcpConfigService))

// ========== 启动 MCP 服务器 ==========
mcpConfig, err := mcp.LoadConfig()
if err != nil {
    log.Printf("加载 MCP 配置失败: %v, 使用默认配置", err)
    mcpConfig = mcp.DefaultConfig()
}

httpMcpServices, err := mcp.BuildDefaultHttpMcpServices()
if err != nil {
    log.Fatalf("构建 HTTP MCP 服务失败: %v", err)
}
// GUI 模式使用已经创建的 gameExcelSvc 实例，保持一致性
httpMcpServices.HttpMcpGameExcelService = gameExcelSvc

mcpServer := mcp.NewMCPServer(mcpConfig, httpMcpServices)

mcpConfigService.SetServer(mcpServer)
```

- [ ] **Step 2: 清理不再需要的 import**

删除 `cmd/rain-qa-func/wails.go` 中因 inline 组装而引入的以下 import（如果后续无其他使用）：

- `"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/robotext"`
- `"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/mcp"`（注意这是 mcputil 别名，检查是否其他代码使用）
- `exceltest`, `functiontest`, `herowikicheck`, `activitywikicheck`, `resourcechecker` 中仅因 inline 组装使用的 import

注意：保留所有其他代码仍在使用的 import。运行 `goimports` 或 `go build` 辅助判断。

- [ ] **Step 3: 编译验证**

Run: `go build ./cmd/rain-qa-func/`
Expected: 编译通过

- [ ] **Step 4: Commit**

```bash
git add cmd/rain-qa-func/wails.go
git commit -m "refactor(wails): use BuildDefaultHttpMcpServices for service assembly

- Replace inline MCP service construction with BuildDefaultHttpMcpServices
- Keep GUI-specific gameExcelSvc override for consistency
- Clean up now-unused imports

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 4: 实现 cobra.go stdio MCP 服务器

**Files:**
- Modify: `backend/pkg/settings/mcp/cobra.go`
- Modify: `backend/pkg/settings/mcp/cobra-help.md`
- Test: `backend/pkg/settings/mcp/cobra_test.go`

**Interfaces:**
- Consumes: `BuildDefaultStdioMcpServices()`（Task 1）
- Produces: `func NewStdioMCPServer(services *StdioMcpServices) *StdioMCPServer`

- [ ] **Step 1: 添加 stdio 服务器实现**

在 `backend/pkg/settings/mcp/cobra.go` 中追加 `StdioMCPServer` 实现：

```go
// StdioMCPServer stdio 传输的 MCP 服务器
type StdioMCPServer struct {
	server   *mcpgo.Server
	services *StdioMcpServices
}

// NewStdioMCPServer 创建 stdio MCP 服务器实例
func NewStdioMCPServer(services *StdioMcpServices) *StdioMCPServer {
	s := mcpgo.NewServer(&mcpgo.Implementation{
		Name:    "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func",
		Version: "1.0.0",
	}, nil)

	return &StdioMCPServer{
		server:   s,
		services: services,
	}
}

// Start 启动 stdio MCP 服务器，从 stdin 读取请求并写入 stdout
func (s *StdioMCPServer) Start(ctx context.Context) error {
	s.registerStdioTools()

	stdioServer := mcpgo.NewStdioServer(s.server)
	return stdioServer.Listen(ctx, os.Stdin, os.Stdout)
}

// registerStdioTools 注册所有 stdio MCP tools
func (s *StdioMCPServer) registerStdioTools() {
	// 注册功能测试相关 tools
	functiontest.RegisterJsonCaseTools(s.server, s.services.StdioMcpJsonCaseService)

	// 注册配置管理 tools
	settings.RegisterConfigTools(s.server, s.services.StdioMcpFuncCaseConfigService, s.services.StdioMcpExcelConfigService, s.services.StdioMcpConfigService)

	// 注册 Excel 检查 tools
	exceltest.RegisterExcelCheckTools(s.server, s.services.StdioMcpExcelCheckService)

	// 注册 Excel 配置 tools
	exceltest.RegisterExcelConfigTools(s.server, s.services.StdioMcpExcelConfigService)

	// 注册 Wiki 检查 tools
	herowikicheck.RegisterWikiCheckTools(s.server, s.services.StdioMcpHeroWikiResCheckService)

	// 注册活动 Wiki 检查 tools
	activitywikicheck.RegisterActivityWikiCheckTools(s.server, s.services.StdioMcpActivityWikiCheckService)

	// 注册游戏数据 tools（Excel 测试页面专用）
	exceltest.RegisterGameExcelTools(s.server, s.services.StdioMcpExcelTestGameService)

	// 注册通用游戏数据 tools（完整版）
	mcputil.RegisterGameExcelTools(s.server, s.services.StdioMcpGameExcelService)

	// 注册 Robot 扩展工具（工会操作）
	mcputil.RegisterRobotGuildTools(s.server, s.services.StdioMcpRobotExtService)

	// 注册战斗测试 tools
	functiontest.RegisterFightCaseTools(s.server, "", func(heroID int, caseName string, filterCases []string) (map[string][]functiontest.LogEntry, error) {
		config, err := s.services.StdioMcpFuncCaseConfigService.GetConfig()
		if err != nil {
			return nil, fmt.Errorf("获取配置失败: %w", err)
		}
		fightCasesDir := config.JsonsDir

		var caseFiles []string
		var filterCaseNames []string

		if caseName != "" {
			filterCaseNames = []string{caseName}
		}

		if len(filterCases) > 0 {
			caseFiles = filterCases
		} else if heroID > 0 {
			pattern1 := filepath.Join(fightCasesDir, fmt.Sprintf("%d_*.json", heroID))
			matches1, _ := filepath.Glob(pattern1)
			pattern2 := filepath.Join(fightCasesDir, fmt.Sprintf("%d-*.json", heroID))
			matches2, _ := filepath.Glob(pattern2)

			caseFiles = append(caseFiles, matches1...)
			caseFiles = append(caseFiles, matches2...)
			for i, match := range caseFiles {
				caseFiles[i] = filepath.Base(match)
			}
		}

		err = s.services.StdioMcpJsonCaseService.RunRobotTest(
			config.ServerAddr,
			fmt.Sprintf("%d", config.ServerPort),
			"stdio-mcp-fight",
			fmt.Sprintf("战斗测试-英雄%d", heroID),
			"",
			5,
			caseFiles,
			filterCaseNames,
			&fightCasesDir,
			nil,
			30000,
			false,
			false,
			false,
			1,
		)
		if err != nil {
			return nil, err
		}

		logs := s.services.StdioMcpJsonCaseService.GetTestLogs(caseName)

		result := make(map[string][]functiontest.LogEntry)
		for caseName, entries := range logs {
			var converted []functiontest.LogEntry
			for _, entry := range entries {
				converted = append(converted, functiontest.LogEntry{
					Case:         entry.Case,
					ID:           entry.ID,
					Level:        entry.Level,
					Type:         entry.Type,
					RobotName:    entry.RobotName,
					Msg:          entry.Msg,
					Time:         entry.Time,
					CodeLocation: entry.CodeLocation,
				})
			}
			result[caseName] = converted
		}

		return result, nil
	}, s.services.StdioMcpFeishuNotifyConfigService)
}
```

- [ ] **Step 2: 修改 NewMCPCmd**

将 `backend/pkg/settings/mcp/cobra.go` 中的 `NewMCPCmd` 改为：

```go
// NewMCPCmd 创建 mcp 子命令。
// 无子命令时启动 stdio 传输的 MCP 服务器。
func NewMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "启动 stdio MCP 服务器",
		Long:  cobraHelpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := BuildDefaultStdioMcpServices()
			if err != nil {
				return fmt.Errorf("构建 stdio MCP 服务失败: %w", err)
			}
			server := NewStdioMCPServer(services)
			return server.Start(cmd.Context())
		},
	}
	return cmd
}
```

- [ ] **Step 3: 添加必要的 import**

在 `backend/pkg/settings/mcp/cobra.go` 顶部添加：

```go
import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	activitywikicheck "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game"
	mcputil "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/mcp"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/robotext"
	exceltest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/excel-test"
	functiontest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test"
	herowikicheck "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/hero-wiki-check"
	resourcechecker "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"
	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)
```

注意：如果后续 `goimports` 发现 `context`、`activitywikicheck`、`game`、`robotext`、`resourcechecker` 未使用（因只在 stdio 注册函数中使用），请保留实际使用到的 import。

- [ ] **Step 4: 修改 cobra-help.md**

将 `backend/pkg/settings/mcp/cobra-help.md` 替换为：

```markdown
# mcp — MCP 服务器管理工具

## 简介

`mcp` 子命令用于启动一个 **stdio 传输** 的内置 MCP（Model Context Protocol）服务器。

当外部 MCP 客户端（如 Claude Code）通过 stdio 启动本程序时，客户端可以通过 `stdin/stdout` 调用项目中的所有 MCP 工具。

## 用法

```bash
rain-qa-func mcp
```

无子命令时直接启动 stdio MCP 服务器。

## 启动方式

外部客户端通常会使用如下方式启动：

```json
{
  "mcpServers": {
    "rain-qa-func": {
      "command": "rain-qa-func",
      "args": ["mcp"]
    }
  }
}
```

## 暴露的工具

stdio 模式下暴露的工具与 GUI 模式下启动的 HTTP MCP 服务器完全一致，包括：

- 功能测试用例管理
- 功能测试配置 / Excel 配置 / MCP 配置
- Excel 检查
- 武将 Wiki 检查
- 活动 Wiki 检查
- 游戏数据查询
- Robot 工会扩展
- 战斗测试

完整工具列表及参数说明见项目文档 `docs/MCP-USAGE.md`。

## 配置行为

stdio 模式**忽略**持久化的 MCP 配置文件，被调用即直接启动。绑定地址、端口等 HTTP 相关配置在 stdio 模式下不适用。

## 输出约定

stdio 模式下所有业务输出均通过 MCP JSON-RPC 通道返回，避免污染 stdout。运行期日志输出到 stderr。

## 示例

```bash
# 启动 stdio MCP 服务器
rain-qa-func mcp

# 查看帮助
rain-qa-func mcp --help
```

## 更多信息

详见 `backend/pkg/settings/mcp/` 目录及 `CLAUDE.md`。
```

- [ ] **Step 5: 编写 cobra_test.go**

创建 `backend/pkg/settings/mcp/cobra_test.go`：

```go
package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStdioMCPServer(t *testing.T) {
	services, err := BuildDefaultStdioMcpServices()
	require.NoError(t, err)

	server := NewStdioMCPServer(services)
	require.NotNil(t, server)
	assert.NotNil(t, server.server)
	assert.NotNil(t, server.services)
}

func TestNewMCPCmd(t *testing.T) {
	cmd := NewMCPCmd()
	require.NotNil(t, cmd)
	assert.Equal(t, "mcp", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}
```

- [ ] **Step 6: 编译验证**

Run: `go build ./backend/pkg/settings/mcp/`
Expected: 编译通过

Run: `go test ./backend/pkg/settings/mcp/ -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add backend/pkg/settings/mcp/cobra.go backend/pkg/settings/mcp/cobra-help.md backend/pkg/settings/mcp/cobra_test.go
git commit -m "feat(mcp): implement stdio MCP server in cobra.go

- Add StdioMCPServer with full tool registration matching HTTP server
- NewMCPCmd starts stdio server using BuildDefaultStdioMcpServices
- Update cobra-help.md for stdio usage
- Add unit tests for stdio server creation and command

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 5: 更新 CLAUDE.md 文档

**Files:**
- Modify: `backend/pkg/settings/mcp/CLAUDE.md`

- [ ] **Step 1: 在目录结构表格中添加 services.go**

将目录结构更新为：

```
mcp/
├── cobra.go        # MCP CLI 子命令（stdio 服务器入口）
├── cobra-help.md   # --help 文档
├── config.go       # MCP 配置加载与持久化
├── server.go       # HTTP MCP 服务器启动、停止、工具注册
├── services.go     # StdioMcpServices / HttpMcpServices 服务容器
├── server_test.go  # 单元测试
├── services_test.go# 单元测试
├── cobra_test.go   # 单元测试
└── tests/          # 手动/集成测试脚本
    └── test_mcp.bat
```

- [ ] **Step 2: 在核心类型表格中添加服务容器**

新增：

| `StdioMcpServices` | [services.go:11](services.go:11) | stdio 模式服务容器 |
| `HttpMcpServices` | [services.go:30](services.go:30) | HTTP 模式服务容器 |

- [ ] **Step 3: 在核心函数表格中添加构建函数**

新增：

| `BuildDefaultStdioMcpServices` | [services.go:80](services.go:80) | 创建 stdio 模式默认服务集合 |
| `BuildDefaultHttpMcpServices` | [services.go:112](services.go:112) | 创建 HTTP 模式默认服务集合 |
| `NewStdioMCPServer` | [cobra.go:35](cobra.go:35) | 创建 stdio MCP 服务器 |
| `StdioMCPServer.Start` | [cobra.go:48](cobra.go:48) | 启动 stdio 服务器 |

- [ ] **Step 4: 添加传输协议说明**

在“传输协议”小节后新增：

```markdown
## CLI stdio 模式

- `rain-qa-func mcp` 无子命令时启动 stdio MCP 服务器。
- stdio 模式忽略 `mcpSection` 配置文件，直接启动并暴露完整工具集。
- 所有业务输出通过 JSON-RPC 返回 stdout；日志输出到 stderr。
```

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/settings/mcp/CLAUDE.md
git commit -m "docs(mcp): update CLAUDE.md for stdio server and service containers

- Document services.go, cobra.go stdio mode, and new core types/functions
- Add CLI stdio mode section

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 6: 集成验证

**Files:**
- 无新增或修改，仅运行验证命令

- [ ] **Step 1: 运行 MCP 包测试**

Run: `go test ./backend/pkg/settings/mcp/ -v`
Expected: 所有测试 PASS

- [ ] **Step 2: 运行整个项目构建**

Run: `go build ./...`
Expected: 编译通过

- [ ] **Step 3: 运行 Wails 构建（如环境可用）**

Run: `wails3 task build`
Expected: 构建成功

- [ ] **Step 4: 手动验证 stdio 命令可执行**

Run: `go run ./cmd/rain-qa-func mcp --help`
Expected: 输出更新后的帮助文档

- [ ] **Step 5: 运行 IDE 诊断**

使用 `mcp__ide__getDiagnostics` 检查所有修改文件的 linting/类型错误。

- [ ] **Step 6: Commit（如验证通过）**

```bash
git commit --allow-empty -m "chore(mcp): integration verification passed

- go test ./backend/pkg/settings/mcp/ PASS
- go build ./... PASS
- wails3 task build PASS
- stdio help output verified

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## 自我审查

### Spec 覆盖检查

| 规格要求 | 实现任务 |
|---|---|
| stdio 传输 MCP 服务器 | Task 4 |
| 完整服务 | Task 1-4 |
| 复用服务初始化逻辑 | Task 1, Task 3 |
| 忽略配置文件 | Task 4（RunE 中不调用 LoadConfig） |
| StdioMcpServices / HttpMcpServices 命名 | Task 1 |
| 更新文档 | Task 5 |
| 测试策略 | Task 1, Task 2, Task 4, Task 6 |

### Placeholder 检查

- 无 "TBD"、"TODO"、"implement later"。
- 所有代码步骤包含完整代码。
- 所有测试步骤包含完整测试代码。
- 所有命令包含预期输出。

### 类型一致性检查

- `BuildDefaultStdioMcpServices()` 返回 `*StdioMcpServices`。
- `BuildDefaultHttpMcpServices()` 返回 `*HttpMcpServices`。
- `NewMCPServer` 参数为 `*HttpMcpServices`。
- `NewStdioMCPServer` 参数为 `*StdioMcpServices`。
- `registerStdioTools` 使用 `s.services.StdioMcpXxx` 字段。
- `registerTools` 使用 `s.services.HttpMcpXxx` 字段。

计划完整，可以开始实现。
