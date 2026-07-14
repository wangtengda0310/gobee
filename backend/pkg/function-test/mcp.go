package functiontest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/mcp"

	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/log_def"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/log_service"

	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterJsonCaseTools 注册功能测试相关的 MCP Tools
// @mcp
func RegisterJsonCaseTools(s *mcpgo.Server, svc *JsonCaseService) {
	// get_case_list - 获取测试用例列表
	s.AddTool(&mcpgo.Tool{
		Name:        "get_case_list",
		Description: "获取测试用例列表，支持文件或目录路径",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filePath": map[string]any{
					"type":        "string",
					"description": "JSON 文件或目录路径",
				},
			},
			"required": []string{"filePath"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			FilePath string `json:"filePath"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		cases, err := svc.GetCaseList(args.FilePath)
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		jsonData, err := json.MarshalIndent(cases, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(jsonData)), nil
	})

	// get_categories - 获取分类信息
	s.AddTool(&mcpgo.Tool{
		Name:        "get_categories",
		Description: "获取分类信息，返回目录下的所有分类及用例",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"dirPath": map[string]any{
					"type":        "string",
					"description": "目录路径",
				},
			},
			"required": []string{"dirPath"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			DirPath string `json:"dirPath"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		categories, err := svc.GetCategories(args.DirPath)
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		jsonData, err := json.MarshalIndent(categories, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(jsonData)), nil
	})

	// search_cases - 搜索用例
	s.AddTool(&mcpgo.Tool{
		Name:        "search_cases",
		Description: "搜索用例，根据关键词在用例名和描述中搜索",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filePath": map[string]any{
					"type":        "string",
					"description": "JSON 文件或目录路径",
				},
				"keyword": map[string]any{
					"type":        "string",
					"description": "搜索关键词",
				},
			},
			"required": []string{"filePath", "keyword"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			FilePath string `json:"filePath"`
			Keyword  string `json:"keyword"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		cases, err := svc.SearchCases(args.FilePath, args.Keyword)
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		jsonData, err := json.MarshalIndent(cases, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(jsonData)), nil
	})

	// run_robot_test - 执行机器人测试
	s.AddTool(&mcpgo.Tool{
		Name:        "run_robot_test",
		Description: "执行机器人测试，启动自动化测试流程",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"ip": map[string]any{
					"type":        "string",
					"description": "服务器 IP 地址",
				},
				"port": map[string]any{
					"type":        "string",
					"description": "服务器端口",
				},
				"prefix": map[string]any{
					"type":        "string",
					"description": "前缀标识",
				},
				"desc": map[string]any{
					"type":        "string",
					"description": "描述信息",
				},
				"feishuGuid": map[string]any{
					"type":        "string",
					"description": "飞书 GUID",
				},
				"loginTime": map[string]any{
					"type":        "number",
					"description": "登录等待时间（秒）",
				},
				"filterCases": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "过滤用例列表",
				},
				"filterCaseNames": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "过滤用例名称列表",
				},
				"dir": map[string]any{
					"type":        "string",
					"description": "用例目录路径（可选）",
				},
				"opTimeMs": map[string]any{
					"type":        "integer",
					"description": "操作超时时间（毫秒）",
				},
				"feishuNtf": map[string]any{
					"type":        "boolean",
					"description": "是否发送飞书通知",
				},
				"debugLevel": map[string]any{
					"type":        "boolean",
					"description": "是否启用调试级别日志",
				},
				"debugLog": map[string]any{
					"type":        "boolean",
					"description": "是否启用调试日志",
				},
				"concurrency": map[string]any{
					"type":        "integer",
					"description": "并发数",
				},
			},
			"required": []string{"ip", "port", "prefix", "desc", "feishuGuid", "loginTime", "filterCases", "filterCaseNames", "opTimeMs", "feishuNtf", "debugLevel", "debugLog", "concurrency"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			IP              string   `json:"ip"`
			Port            string   `json:"port"`
			Prefix          string   `json:"prefix"`
			Desc            string   `json:"desc"`
			FeishuGuid      string   `json:"feishuGuid"`
			LoginTime       float64  `json:"loginTime"`
			FilterCases     []string `json:"filterCases"`
			FilterCaseNames []string `json:"filterCaseNames"`
			Dir             *string  `json:"dir"`
			OpTimeMs        int      `json:"opTimeMs"`
			FeishuNtf       bool     `json:"feishuNtf"`
			DebugLevel      bool     `json:"debugLevel"`
			DebugLog        bool     `json:"debugLog"`
			Concurrency     uint     `json:"concurrency"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		// 执行测试
		err := svc.RunRobotTest(
			args.IP, args.Port, args.Prefix, args.Desc, args.FeishuGuid,
			args.LoginTime,
			args.FilterCases, args.FilterCaseNames,
			args.Dir,
			nil, // filesData - MCP 不支持文件数据传输
			args.OpTimeMs,
			args.FeishuNtf, args.DebugLevel, args.DebugLog,
			args.Concurrency,
		)

		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult("测试执行完成"), nil
	})

	// stop_robot_test - 停止测试
	s.AddTool(&mcpgo.Tool{
		Name:        "stop_robot_test",
		Description: "停止正在运行的机器人测试",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		err := svc.StopRobotTest()
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult("测试已停止"), nil
	})

	// is_running - 检查是否在运行
	s.AddTool(&mcpgo.Tool{
		Name:        "is_running",
		Description: "检查机器人测试是否正在运行",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		running := svc.IsRunningRobotTest()

		result := map[string]bool{
			"running": running,
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(jsonData)), nil
	})

	// get_test_logs - 获取测试日志
	s.AddTool(&mcpgo.Tool{
		Name:        "get_test_logs",
		Description: "获取测试运行日志，支持按用例名过滤",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"caseName": map[string]any{
					"type":        "string",
					"description": "用例名称（可选，不填则返回所有日志）",
				},
			},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			CaseName string `json:"caseName"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		logs := svc.GetTestLogs(args.CaseName)

		jsonData, err := json.MarshalIndent(logs, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(jsonData)), nil
	})

	// clear_test_logs - 清除测试日志
	s.AddTool(&mcpgo.Tool{
		Name:        "clear_test_logs",
		Description: "清除缓存的测试日志",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		svc.ClearTestLogs()
		return mcp.TextResult("日志已清除"), nil
	})

	// get_case_by_name - 获取单个用例的完整数据
	s.AddTool(&mcpgo.Tool{
		Name:        "get_case_by_name",
		Description: "获取单个测试用例的完整数据，包括 steps、initYanWu、customHeroes 等详细信息",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filePath": map[string]any{
					"type":        "string",
					"description": "JSON 文件或目录路径",
				},
				"caseName": map[string]any{
					"type":        "string",
					"description": "用例名称",
				},
			},
			"required": []string{"filePath", "caseName"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			FilePath string `json:"filePath"`
			CaseName string `json:"caseName"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		testCase, err := svc.GetTestCase(args.FilePath, args.CaseName)
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		if testCase == nil {
			return mcp.ErrorResult("未找到用例: " + args.CaseName), nil
		}

		jsonData, err := json.MarshalIndent(testCase, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(jsonData)), nil
	})
}

// FightCaseInfo 战斗用例信息
type FightCaseInfo struct {
	File     string `json:"file"`
	CaseName string `json:"caseName"`
	Desc     string `json:"desc"`
	HeroID   int    `json:"heroId,omitempty"`
}

// FightTestResult 战斗测试结果
type FightTestResult struct {
	Success bool                  `json:"success"`
	Message string                `json:"message"`
	Logs    map[string][]LogEntry `json:"logs,omitempty"`
}

// FightTestSummary 战斗测试结果汇总
type FightTestSummary struct {
	TotalCases  int          `json:"totalCases"`
	Passed      int          `json:"passed"`
	Failed      int          `json:"failed"`
	FailedCases []FailedCase `json:"failedCases,omitempty"`
}

// FailedCase 失败用例信息
type FailedCase struct {
	CaseName string `json:"caseName"`
	Error    string `json:"error"`
}

// AsyncTestResult 异步测试结果
type AsyncTestResult struct {
	TaskID   string `json:"taskId"`   // 任务ID，用于查询进度
	Status   string `json:"status"`   // pending, running, completed, failed
	Progress string `json:"progress"` // 进度描述，如 "3/17"
}

// TestProgress 测试进度
type TestProgress struct {
	TaskID       string            `json:"taskId"`
	Status       string            `json:"status"` // pending, running, completed, failed
	TotalCases   int               `json:"totalCases"`
	Completed    int               `json:"completed"`
	CurrentCase  string            `json:"currentCase,omitempty"`
	FailedCases  []string          `json:"failedCases,omitempty"`
	Summary      *FightTestSummary `json:"summary,omitempty"` // 完成后填充
	ErrorMessage string            `json:"errorMessage,omitempty"`
}

// TaskStatus 任务状态常量
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)

// taskManager 任务管理器，维护所有异步任务的状态
type taskManager struct {
	mu     sync.RWMutex
	tasks  map[string]*TestProgress
	nextID int
}

// globalTaskManager 全局任务管理器实例
var globalTaskManager = &taskManager{
	tasks: make(map[string]*TestProgress),
}

// createTask 创建新任务
func (tm *taskManager) createTask() *TestProgress {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.nextID++
	taskID := fmt.Sprintf("task_%d_%d", time.Now().Unix(), tm.nextID)

	task := &TestProgress{
		TaskID:      taskID,
		Status:      TaskStatusPending,
		FailedCases: []string{},
	}
	tm.tasks[taskID] = task
	return task
}

// getTask 获取任务
func (tm *taskManager) getTask(taskID string) (*TestProgress, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	task, ok := tm.tasks[taskID]
	return task, ok
}

// updateTask 更新任务
func (tm *taskManager) updateTask(taskID string, update func(*TestProgress)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if task, ok := tm.tasks[taskID]; ok {
		update(task)
	}
}

// FightCaseService 战斗用例服务接口
type FightCaseService interface {
	GetFightCases(dir string) ([]FightCaseInfo, error)
	RunFightTest(heroID int, caseName string) error
}

// FightCaseTools 战斗测试工具集合
type FightCaseTools struct {
	fightCasesDir         string
	runTestFunc           func(heroID int, caseName string, filterCases []string) (map[string][]LogEntry, error)
	feishuNotifyConfigSvc feishu.FeishuNotifier
}

// NewFightCaseTools 创建战斗测试工具实例
func NewFightCaseTools(fightCasesDir string, runTestFunc func(heroID int, caseName string, filterCases []string) (map[string][]LogEntry, error), feishuNotifyConfigSvc feishu.FeishuNotifier) *FightCaseTools {
	return &FightCaseTools{
		fightCasesDir:         fightCasesDir,
		runTestFunc:           runTestFunc,
		feishuNotifyConfigSvc: feishuNotifyConfigSvc,
	}
}

// RegisterFightCaseTools 注册战斗测试相关的 MCP Tools
// @mcp
func RegisterFightCaseTools(s *mcpgo.Server, fightCasesDir string, runTestFunc func(heroID int, caseName string, filterCases []string) (map[string][]LogEntry, error), feishuNotifyConfigSvc feishu.FeishuNotifier) {
	tools := NewFightCaseTools(fightCasesDir, runTestFunc, feishuNotifyConfigSvc)

	// get_fight_cases - 获取战斗用例列表
	s.AddTool(&mcpgo.Tool{
		Name:        "get_fight_cases",
		Description: "获取战斗测试用例列表，支持按英雄ID或名称过滤",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"heroId": map[string]any{
					"type":        "integer",
					"description": "英雄ID（可选）",
				},
				"keyword": map[string]any{
					"type":        "string",
					"description": "搜索关键词（可选）",
				},
			},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			HeroID  *int   `json:"heroId"`
			Keyword string `json:"keyword"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		cases, err := getFightCases(fightCasesDir, args.HeroID, args.Keyword)
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		jsonData, err := json.MarshalIndent(cases, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(jsonData)), nil
	})

	// run_fight_test - 运行战斗测试
	s.AddTool(&mcpgo.Tool{
		Name:        "run_fight_test",
		Description: "运行指定英雄的战斗测试（如果配置了飞书自动通知，测试完成后会自动发送通知）",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"heroId": map[string]any{
					"type":        "integer",
					"description": "英雄ID",
				},
				"caseName": map[string]any{
					"type":        "string",
					"description": "用例名称（可选，不填则运行该英雄所有用例）",
				},
				"filterCases": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "指定运行的用例文件名列表（可选）",
				},
			},
			"required": []string{"heroId"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			HeroID      int      `json:"heroId"`
			CaseName    string   `json:"caseName"`
			FilterCases []string `json:"filterCases"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		// 获取英雄名称
		heroName := tools.getHeroNameByTools(args.HeroID)

		// 执行测试
		logs, err := runTestFunc(args.HeroID, args.CaseName, args.FilterCases)

		result := FightTestResult{
			Success: err == nil,
			Message: "战斗测试完成",
			Logs:    logs,
		}

		if err != nil {
			result.Message = "战斗测试失败: " + err.Error()
		}

		// 检查是否需要发送飞书通知
		if tools.checkAndSendFeishuNotify(heroName, result) {
			result.Message += " (已发送飞书通知)"
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(jsonData)), nil
	})

	// get_hero_list - 获取有测试用例的英雄列表
	s.AddTool(&mcpgo.Tool{
		Name:        "get_hero_list",
		Description: "获取有战斗测试用例的英雄列表",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		heroes, err := getHeroList(fightCasesDir)
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		jsonData, err := json.MarshalIndent(heroes, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(jsonData)), nil
	})

	// get_fight_test_summary - 获取战斗测试结果汇总
	s.AddTool(&mcpgo.Tool{
		Name:        "get_fight_test_summary",
		Description: "获取战斗测试结果汇总，返回简洁的测试统计（通过/失败数量、失败用例错误信息）",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		summary := getFightTestSummary()

		jsonData, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(jsonData)), nil
	})

	// run_fight_test_async - 异步运行战斗测试
	s.AddTool(&mcpgo.Tool{
		Name:        "run_fight_test_async",
		Description: "异步运行指定英雄的战斗测试，立即返回任务ID，可通过 get_test_progress 查询进度",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"heroId": map[string]any{
					"type":        "integer",
					"description": "英雄ID",
				},
				"caseName": map[string]any{
					"type":        "string",
					"description": "用例名称（可选，不填则运行该英雄所有用例）",
				},
				"filterCases": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "指定运行的用例文件名列表（可选）",
				},
			},
			"required": []string{"heroId"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			HeroID      int      `json:"heroId"`
			CaseName    string   `json:"caseName"`
			FilterCases []string `json:"filterCases"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		// 创建任务
		task := globalTaskManager.createTask()

		// 获取要执行的用例数量，用于进度显示
		cases, err := getFightCases(fightCasesDir, &args.HeroID, "")
		if err != nil {
			globalTaskManager.updateTask(task.TaskID, func(t *TestProgress) {
				t.Status = TaskStatusFailed
				t.ErrorMessage = "获取用例列表失败: " + err.Error()
			})
			return mcp.ErrorResult("获取用例列表失败: " + err.Error()), nil
		}
		globalTaskManager.updateTask(task.TaskID, func(t *TestProgress) {
			t.TotalCases = len(cases)
		})

		// 启动异步执行
		go runFightTestAsync(task.TaskID, args.HeroID, args.CaseName, args.FilterCases, runTestFunc)

		// 立即返回任务信息
		result := AsyncTestResult{
			TaskID:   task.TaskID,
			Status:   TaskStatusPending,
			Progress: fmt.Sprintf("0/%d", len(cases)),
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(jsonData)), nil
	})

	// get_test_progress - 查询测试进度
	s.AddTool(&mcpgo.Tool{
		Name:        "get_test_progress",
		Description: "查询异步测试任务的执行进度",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"taskId": map[string]any{
					"type":        "string",
					"description": "任务ID（由 run_fight_test_async 返回）",
				},
			},
			"required": []string{"taskId"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			TaskID string `json:"taskId"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		task, ok := globalTaskManager.getTask(args.TaskID)
		if !ok {
			return mcp.ErrorResult("任务不存在: " + args.TaskID), nil
		}

		jsonData, err := json.MarshalIndent(task, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(jsonData)), nil
	})

	// get_feishu_notify_config - 获取飞书自动通知配置
	s.AddTool(&mcpgo.Tool{
		Name:        "get_feishu_notify_config",
		Description: "获取战斗测试的飞书自动通知配置",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, tools.handleGetFeishuNotifyConfig)

	// set_feishu_notify_config - 设置飞书自动通知配置
	s.AddTool(&mcpgo.Tool{
		Name:        "set_feishu_notify_config",
		Description: "设置战斗测试的飞书自动通知配置（测试完成后自动发送通知）",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"enabled": map[string]any{
					"type":        "boolean",
					"description": "是否启用飞书自动通知",
				},
				"robotGuid": map[string]any{
					"type":        "string",
					"description": "飞书机器人GUID",
				},
				"messageTemplate": map[string]any{
					"type":        "string",
					"description": "消息模板，支持变量: {heroName}, {total}, {passed}, {failed}, {passRate}",
				},
			},
			"required": []string{"enabled", "robotGuid"},
		},
	}, tools.handleSetFeishuNotifyConfig)
}

// getFightCases 获取战斗用例列表
func getFightCases(dir string, heroID *int, keyword string) ([]FightCaseInfo, error) {
	var cases []FightCaseInfo

	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileName := filepath.Base(file)

		// 从文件名提取英雄ID (格式: 10103_关羽.json 或 10112-魏延.json)
		var fileHeroID int
		// 将连字符替换为下划线，统一处理
		normalizedName := strings.Replace(fileName, "-", "_", 1)
		parts := strings.Split(normalizedName, "_")
		if len(parts) >= 1 {
			if _, err := json.Number(parts[0]).Int64(); err == nil {
				fileHeroID = int(mustParseInt(parts[0]))
			}
		}

		// 过滤英雄ID
		if heroID != nil && fileHeroID != *heroID {
			continue
		}

		// 读取文件内容获取用例信息
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var caseList []map[string]any
		if err := json.Unmarshal(data, &caseList); err != nil {
			continue
		}

		for _, c := range caseList {
			caseName, _ := c["case"].(string)
			desc, _ := c["desc"].(string)

			// 关键词过滤
			if keyword != "" {
				if !strings.Contains(caseName, keyword) && !strings.Contains(desc, keyword) && !strings.Contains(fileName, keyword) {
					continue
				}
			}

			cases = append(cases, FightCaseInfo{
				File:     fileName,
				CaseName: caseName,
				Desc:     desc,
				HeroID:   fileHeroID,
			})
		}
	}

	return cases, nil
}

// HeroInfo 英雄信息
type HeroInfo struct {
	HeroID   int    `json:"heroId"`
	Name     string `json:"name"`
	FileName string `json:"fileName"`
	CaseNum  int    `json:"caseNum"`
}

// getHeroList 获取有测试用例的英雄列表
func getHeroList(dir string) ([]HeroInfo, error) {
	heroMap := make(map[int]*HeroInfo)

	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileName := filepath.Base(file)

		// 从文件名提取英雄ID和名称（支持下划线和连字符两种格式）
		baseName := strings.TrimSuffix(fileName, ".json")
		// 将连字符替换为下划线，统一处理
		normalizedName := strings.Replace(baseName, "-", "_", 1)
		parts := strings.Split(normalizedName, "_")
		if len(parts) < 2 {
			// 非英雄文件 (如: 1_强化行动.json)
			continue
		}

		heroID := mustParseInt(parts[0])
		if heroID == 0 {
			continue
		}

		heroName := strings.Join(parts[1:], "_")

		// 统计用例数量
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var caseList []map[string]any
		if err := json.Unmarshal(data, &caseList); err != nil {
			continue
		}

		if existing, ok := heroMap[heroID]; ok {
			existing.CaseNum += len(caseList)
		} else {
			heroMap[heroID] = &HeroInfo{
				HeroID:   heroID,
				Name:     heroName,
				FileName: fileName,
				CaseNum:  len(caseList),
			}
		}
	}

	// 转换为列表
	result := make([]HeroInfo, 0, len(heroMap))
	for _, h := range heroMap {
		result = append(result, *h)
	}

	return result, nil
}

func mustParseInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			break
		}
	}
	return result
}

// getFightTestSummary 从日志缓存中解析测试结果汇总
func getFightTestSummary() FightTestSummary {
	allLogs := log_service.GetAllLogs()

	summary := FightTestSummary{
		TotalCases:  len(allLogs),
		FailedCases: make([]FailedCase, 0),
	}

	for caseName, logs := range allLogs {
		hasError := false
		var errorMsgs []string

		for _, log := range logs {
			if log.Level == log_def.ERROR {
				hasError = true
				errorMsgs = append(errorMsgs, log.Msg)
			}
		}

		if hasError {
			summary.Failed++
			// 合并所有错误信息，取第一条（通常是最关键的）
			errMsg := "测试失败"
			if len(errorMsgs) > 0 {
				// 限制错误信息长度，避免过长
				if len(errorMsgs[0]) > 200 {
					errMsg = errorMsgs[0][:200] + "..."
				} else {
					errMsg = errorMsgs[0]
				}
			}
			summary.FailedCases = append(summary.FailedCases, FailedCase{
				CaseName: caseName,
				Error:    errMsg,
			})
		} else {
			summary.Passed++
		}
	}

	return summary
}

// handleGetFeishuNotifyConfig 处理获取飞书通知配置请求
func (t *FightCaseTools) handleGetFeishuNotifyConfig(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	config, err := t.feishuNotifyConfigSvc.GetConfig()
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	return mcp.TextResult(string(data)), nil
}

// handleSetFeishuNotifyConfig 处理设置飞书通知配置请求
func (t *FightCaseTools) handleSetFeishuNotifyConfig(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	var args struct {
		Enabled         bool   `json:"enabled"`
		RobotGUID       string `json:"robotGuid"`
		MessageTemplate string `json:"messageTemplate"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
	}

	config := &feishu.FeishuNotifyConfig{
		Enabled:         args.Enabled,
		RobotGUID:       args.RobotGUID,
		MessageTemplate: args.MessageTemplate,
	}

	if err := t.feishuNotifyConfigSvc.SaveConfig(config); err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	result := map[string]any{
		"success": true,
		"message": "飞书通知配置保存成功",
		"config":  config,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	return mcp.TextResult(string(data)), nil
}

// getHeroNameByTools 根据英雄ID获取英雄名称
func (t *FightCaseTools) getHeroNameByTools(heroID int) string {
	heroes, err := getHeroList(t.fightCasesDir)
	if err != nil {
		return fmt.Sprintf("英雄%d", heroID)
	}
	for _, h := range heroes {
		if h.HeroID == heroID {
			return h.Name
		}
	}
	return fmt.Sprintf("英雄%d", heroID)
}

// checkAndSendFeishuNotify 检查并发送飞书通知
func (t *FightCaseTools) checkAndSendFeishuNotify(heroName string, result FightTestResult) bool {
	// 获取飞书通知配置
	config, err := t.feishuNotifyConfigSvc.GetConfig()
	if err != nil {
		return false
	}

	// 检查是否启用
	if !config.Enabled {
		return false
	}

	// 检查机器人GUID是否配置
	if config.RobotGUID == "" {
		return false
	}

	// 统计测试结果
	total, passed, failed := countTestResults(result.Logs)

	// 计算通过率
	passRate := 0.0
	if total > 0 {
		passRate = float64(passed) / float64(total) * 100
	}

	// 构建消息
	message := buildFeishuMessage(config.MessageTemplate, heroName, total, passed, failed, passRate)

	// 发送消息
	feishu.SendFeiShuRobotText(config.RobotGUID, "%s", message)

	return true
}

// countTestResults 统计测试结果
func countTestResults(logs map[string][]LogEntry) (total, passed, failed int) {
	for _, caseLogs := range logs {
		hasError := false
		for _, log := range caseLogs {
			// 根据日志级别判断测试结果
			// ERROR 级别表示失败
			if log.Level == "ERROR" || log.Level == "FAIL" {
				hasError = true
			}
		}
		total++
		if hasError {
			failed++
		} else {
			passed++
		}
	}

	return total, passed, failed
}

// buildFeishuMessage 构建飞书消息
func buildFeishuMessage(template, heroName string, total, passed, failed int, passRate float64) string {
	// 替换模板变量
	message := template
	message = strings.ReplaceAll(message, "{heroName}", heroName)
	message = strings.ReplaceAll(message, "{total}", fmt.Sprintf("%d", total))
	message = strings.ReplaceAll(message, "{passed}", fmt.Sprintf("%d", passed))
	message = strings.ReplaceAll(message, "{failed}", fmt.Sprintf("%d", failed))
	message = strings.ReplaceAll(message, "{passRate}", fmt.Sprintf("%.1f", passRate))

	return message
}

// runFightTestAsync 异步执行战斗测试
func runFightTestAsync(taskID string, heroID int, caseName string, filterCases []string, runTestFunc func(heroID int, caseName string, filterCases []string) (map[string][]LogEntry, error)) {
	// 更新状态为运行中
	globalTaskManager.updateTask(taskID, func(t *TestProgress) {
		t.Status = TaskStatusRunning
	})

	// 输出开始日志
	fmt.Println("\n========================================")
	fmt.Println("战斗测试开始执行")
	fmt.Printf("任务ID: %s\n", taskID)
	fmt.Printf("英雄ID: %d\n", heroID)
	if caseName != "" {
		fmt.Printf("指定用例: %s\n", caseName)
	}
	if len(filterCases) > 0 {
		fmt.Printf("过滤用例: %v\n", filterCases)
	}
	fmt.Println("========================================")

	// 执行测试
	logs, err := runTestFunc(heroID, caseName, filterCases)

	// 更新最终状态
	globalTaskManager.updateTask(taskID, func(t *TestProgress) {
		if err != nil {
			t.Status = TaskStatusFailed
			t.ErrorMessage = err.Error()
			return
		}

		t.Status = TaskStatusCompleted
		t.Completed = len(logs)

		// 统计通过和失败的用例
		var passedCount, failedCount int
		var failedCases []string
		for caseName, entries := range logs {
			hasError := false
			for _, entry := range entries {
				if entry.Level == "ERROR" || entry.Level == "FAIL" {
					hasError = true
					break
				}
			}
			if hasError {
				failedCount++
				failedCases = append(failedCases, caseName)
			} else {
				passedCount++
			}
		}

		t.FailedCases = failedCases
		t.Summary = &FightTestSummary{
			TotalCases:  len(logs),
			Passed:      passedCount,
			Failed:      failedCount,
			FailedCases: []FailedCase{}, // 详细失败信息需要从日志解析
		}

		// 输出测试结果日志
		printFightTestResultLog(t, logs)
	})
}

// printFightTestResultLog 输出战斗测试结果日志到命令行
func printFightTestResultLog(progress *TestProgress, logs map[string][]LogEntry) {
	fmt.Println("\n========================================")
	fmt.Println("战斗测试结果汇总")
	fmt.Printf("任务ID: %s\n", progress.TaskID)
	fmt.Printf("完成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("========================================")

	// 统计信息
	fmt.Printf("\n测试统计:\n")
	fmt.Printf("  • 总用例数: %d\n", progress.TotalCases)
	fmt.Printf("  • 通过: %d\n", progress.Summary.Passed)
	fmt.Printf("  • 失败: %d\n", progress.Summary.Failed)

	// 计算通过率
	if progress.TotalCases > 0 {
		passRate := float64(progress.Summary.Passed) / float64(progress.TotalCases) * 100
		fmt.Printf("  • 通过率: %.1f%%\n", passRate)
	}

	// 失败用例详情
	if len(progress.FailedCases) > 0 {
		fmt.Println("\n失败用例:")
		for _, caseName := range progress.FailedCases {
			fmt.Printf("  • %s\n", caseName)

			// 输出该用例的错误日志
			if entries, ok := logs[caseName]; ok {
				for _, entry := range entries {
					if entry.Level == "ERROR" || entry.Level == "FAIL" {
						fmt.Printf("      [%s] %s\n", entry.Level, entry.Msg)
						if entry.CodeLocation != "" {
							fmt.Printf("      位置: %s\n", entry.CodeLocation)
						}
					}
				}
			}
		}
	}

	// 最终结果
	fmt.Println("\n----------------------------------------")
	if progress.Summary.Failed == 0 {
		fmt.Println("所有测试通过！")
	} else {
		fmt.Printf("测试完成，%d 个用例失败\n", progress.Summary.Failed)
	}
	fmt.Println("========================================")
}

// RegisterGameExcelTools 注册游戏数据相关的 MCP Tools
// @mcp
func RegisterGameExcelTools(s *mcpgo.Server, svc *FunctionTestGameService) {
	// get_all_hero_cfg - 获取所有英雄配置
	s.AddTool(&mcpgo.Tool{
		Name:        "get_all_hero_cfg",
		Description: "获取所有英雄配置",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		heroCfg := svc.GetAllHeroCfg()
		data, err := json.Marshal(heroCfg)
		if err != nil {
			return mcp.ErrorResult("序列化英雄配置失败: " + err.Error()), nil
		}
		return mcp.TextResult(string(data)), nil
	})

	// get_all_card_cfg - 获取所有卡牌配置
	s.AddTool(&mcpgo.Tool{
		Name:        "get_all_card_cfg",
		Description: "获取所有卡牌配置",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		cardCfg := svc.GetAllCardCfg()
		data, err := json.Marshal(cardCfg)
		if err != nil {
			return mcp.ErrorResult("序列化卡牌配置失败: " + err.Error()), nil
		}
		return mcp.TextResult(string(data)), nil
	})

	// get_all_skill_cfg - 获取所有技能配置
	s.AddTool(&mcpgo.Tool{
		Name:        "get_all_skill_cfg",
		Description: "获取所有技能配置",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		skillCfg := svc.GetAllSkillCfg()
		data, err := json.Marshal(skillCfg)
		if err != nil {
			return mcp.ErrorResult("序列化技能配置失败: " + err.Error()), nil
		}
		return mcp.TextResult(string(data)), nil
	})

	// get_msg_id_map - 获取消息 ID 映射
	s.AddTool(&mcpgo.Tool{
		Name:        "get_msg_id_map",
		Description: "获取消息 ID 映射 (EGameMsgId 名称映射)",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		msgIdMap := svc.GetEGameMsgIdNameMap()
		data, err := json.Marshal(msgIdMap)
		if err != nil {
			return mcp.ErrorResult("序列化消息ID映射失败: " + err.Error()), nil
		}
		return mcp.TextResult(string(data)), nil
	})

	// get_error_code_map - 获取错误码映射
	s.AddTool(&mcpgo.Tool{
		Name:        "get_error_code_map",
		Description: "获取错误码映射 (ErrorCode 名称映射)",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		errorCodeMap := svc.GetErrorCodeMap()
		data, err := json.Marshal(errorCodeMap)
		if err != nil {
			return mcp.ErrorResult("序列化错误码映射失败: " + err.Error()), nil
		}
		return mcp.TextResult(string(data)), nil
	})

	// get_property_type_map - 获取属性类型映射
	s.AddTool(&mcpgo.Tool{
		Name:        "get_property_type_map",
		Description: "获取属性类型映射 (PropertyType 名称映射，已去除 Pro_ 前缀)",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		propertyTypeMap := svc.GetPropertyTypeMap()
		data, err := json.Marshal(propertyTypeMap)
		if err != nil {
			return mcp.ErrorResult("序列化属性类型映射失败: " + err.Error()), nil
		}
		return mcp.TextResult(string(data)), nil
	})
}
