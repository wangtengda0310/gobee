package prototest

import (
	"context"
	"fmt"

	"github.com/wangtengda0310/gobee/agent/pkg/tool"
)

var llmRecordSvc *RecordControlService

// InitLLMTools 初始化 LLM 工具注册（由 ChatService 调用）
// 接受各独立 Service，替代旧的 StreamProxyService
func InitLLMTools(registry *tool.Registry, recordSvc *RecordControlService, testCaseSvc *TestCaseService, recordFileSvc *RecordFileService) {
	llmRecordSvc = recordSvc

	if registry == nil || recordSvc == nil || testCaseSvc == nil || recordFileSvc == nil {
		return
	}

	registry.MustRegister(
		tool.NewFunction("start_record", "启动协议录制，开始拦截客户端与游戏服务器之间的协议通信",
			func(ctx context.Context, args map[string]any) (any, error) {
				serverAddr, _ := args["serverAddr"].(string)
				httpAddr, _ := args["httpAddr"].(string)
				err := recordSvc.Start(serverAddr, httpAddr, false)
				if err != nil {
					return nil, err
				}
				return "录制已启动", nil
			},
			tool.WithStringParam("serverAddr", "目标 TCP 服务器地址（如 10.0.0.1:18000）", true),
			tool.WithStringParam("httpAddr", "目标 HTTP 地址（如 10.0.0.1:20144）", false),
		),

		tool.NewFunction("stop_record", "停止正在进行的协议录制",
			func(ctx context.Context, args map[string]any) (any, error) {
				_ = recordSvc.StopRecord()
				return "录制已停止", nil
			},
		),

		tool.NewFunction("load_case_list", "获取所有已保存的测试用例列表",
			func(ctx context.Context, args map[string]any) (any, error) {
				return testCaseSvc.LoadTestCaseList()
			},
		),

		tool.NewFunction("load_case", "加载指定测试用例的协议消息数据",
			func(ctx context.Context, args map[string]any) (any, error) {
				name, _ := args["name"].(string)
				return testCaseSvc.LoadTestCase(name)
			},
			tool.WithStringParam("name", "测试用例名称", true),
		),

		tool.NewFunction("delete_case", "删除指定的测试用例",
			func(ctx context.Context, args map[string]any) (any, error) {
				name, _ := args["name"].(string)
				if err := testCaseSvc.DeleteTestCase(name); err != nil {
					return nil, err
				}
				return "已删除: " + name, nil
			},
			tool.WithStringParam("name", "要删除的测试用例名称", true),
		),

		tool.NewFunction("save_case", "将当前录制的协议消息保存为测试用例",
			func(ctx context.Context, args map[string]any) (any, error) {
				name, _ := args["name"].(string)
				// 从内存获取当前录制数据（录制不再自动落盘）
				rec := llmRecordSvc.GetRecording()
				if rec == nil || len(rec.Messages) == 0 {
					return nil, fmt.Errorf("无录制数据，请先录制协议消息")
				}
				views := recordingToViews(rec)
				data := &RecordFileData{
					Version:      rec.Version,
					RecordedAt:   rec.RecordedAt,
					ServerAddr:   rec.ServerAddr,
					MessageCount: len(views),
					Messages:     views,
				}
				if err := testCaseSvc.SaveTestCase(name, data); err != nil {
					return nil, err
				}
				return "已保存为用例: " + name, nil
			},
			tool.WithStringParam("name", "测试用例名称", true),
		),
	)
}
