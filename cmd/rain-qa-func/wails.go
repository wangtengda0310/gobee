package cmd

// 本文件包含 Wails GUI 应用的完整启动逻辑。
// 从原 main.go 迁移而来，由 RootCmd.Run 在无子命令时调用。

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	activitywikicheck "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game"
	exceltest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/excel-test"
	functiontest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test"
	herowikicheck "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/hero-wiki-check"
	prototest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test"
	protocol "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
	serverconfig "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/server-config"
	resourcechecker "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings/home"
	mcp "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings/mcp"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings/roadmap"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings/serverlog"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func init() {
	// 注册 time 事件，绑定生成器会据此生成前端强类型事件 API。
	application.RegisterEvent[string]("time")
}

// getCDPPort 从环境变量 CDP_PORT 读取远程调试端口，默认 9223
func getCDPPort() string {
	port := os.Getenv("CDP_PORT")
	if port == "" {
		port = "9223"
	}
	if _, err := strconv.Atoi(port); err != nil {
		log.Printf("无效的 CDP_PORT 环境变量: %s, 使用默认端口 9223", port)
		port = "9223"
	}
	return port
}

// runWails 启动 Wails GUI 应用。
// 包含完整的服务初始化、窗口创建和事件循环。
func runWails() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6061", nil))
	}()

	// 创建配置服务（需在其他服务之前创建）
	wailsFuncConfigService := functiontest.NewFuncCaseConfigService()

	// 创建飞书消息劫持服务
	interceptService := feishu.NewInterceptService()

	// 创建服务端日志服务（用于拦截 stdout/stderr 并转发到前端）
	serverLogService := serverlog.NewServerLogService()

	// 创建 Excel 检查服务（保存引用，稍后设置事件回调）
	excelCheckService := exceltest.NewExcelCheckServiceWithConfig(wailsFuncConfigService)

	// 创建游戏 Excel 服务并预初始化（从配置读取资源路径）
	gameExcelSvc := game.NewGameExcelService()
	log.Printf("[main] 创建 gameExcelSvc, IsInitialized=%v", gameExcelSvc.IsInitialized())
	if funcConfig, err := wailsFuncConfigService.GetConfig(); err == nil && funcConfig != nil && funcConfig.ExcelResourcesDir != "" {
		log.Printf("[main] 从配置读取Excel路径: %s", funcConfig.ExcelResourcesDir)
		initErr := gameExcelSvc.InitExcel(funcConfig.ExcelResourcesDir)
		log.Printf("[main] InitExcel 结果: err=%v, IsInitialized=%v", initErr, gameExcelSvc.IsInitialized())
	} else {
		log.Printf("[main] 未找到Excel配置路径, err=%v, funcConfig=%v", err, funcConfig)
	}

	// 创建统一策划配表目录配置服务（供 stream-proxy 等模块使用）
	excelConfigSvc := settings.NewExcelConfigService()

	// 创建技能描述服务并注入统一策划配表目录提供者（打破 game↔settings 循环依赖）
	skillUIDescSvc := game.NewSkillUIDescService()
	skillUIDescSvc.SetExcelDirProvider(&excelDirAdapter{svc: excelConfigSvc})

	app := application.New(application.Options{
		Name:        "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(wailsFuncConfigService),
			application.NewService(exceltest.NewExcelConfigService()),
			application.NewService(excelConfigSvc),
			application.NewService(herowikicheck.NewHeroWikiConfigService()),
			application.NewService(activitywikicheck.NewActivityWikiConfigService()),
			application.NewService(gameExcelSvc),
			application.NewService(skillUIDescSvc),
			application.NewService(excelCheckService),
			application.NewService(resourcechecker.NewHeroResCheckService()),
			application.NewService(herowikicheck.NewHeroWikiResCheckService()),
			application.NewService(activitywikicheck.NewActivityWikiCheckService()),
			application.NewService(settings.NewVersionService()),
			application.NewService(interceptService),
		},
		Assets: application.AssetOptions{
			// 使用根目录嵌入的前端资源（由 main.go 通过 SetAssets() 设置）
			Handler: application.AssetFileServerFS(AssetsFS),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Windows: application.WindowsOptions{
			AdditionalBrowserArgs: []string{"--remote-debugging-port=" + getCDPPort()},
		},
		LogLevel: slog.LevelWarn,
	})

	app.RegisterService(application.NewService(functiontest.NewJsonCaseService(app.Event)))

	// 初始化服务端日志服务（替换 stdout/stderr，启动日志转发）
	serverLogService.InitWithApp(app)
	app.RegisterService(application.NewService(serverLogService))

	// 设置 Excel 检查服务的事件发送函数（用于劫持消息推送到前端）
	excelCheckService.SetEventsEmit(func(name string, data any) {
		app.Event.Emit(name, data)
	})

	// 注册协议重放服务（需在聊天服务之前创建，以便注入 LLM 工具）
	recordFileSvc := prototest.NewRecordFileService()

	// 创建全局连接池（所有重放/重发操作共享）
	connPool := protocol.NewAccountConnectionPool()

	// 创建 ReplayWorker 并设置重放/发送函数
	replayWorker := prototest.NewReplayWorker(app.Event)
	replayWorker.SetConnPoolFactory(func() *protocol.AccountConnectionPool {
		return connPool
	})
	replayWorker.SetReplayFunc(func(filePath, serverAddr, httpAddr, openID string, repeatCount int, onProgress func(total, sent int, currentMsg string) bool) error {
		return protocol.Replay(filePath, serverAddr, httpAddr, openID, onProgress, replayWorker.EmitReplayMessage)
	})
	replayWorker.SetSendFunc(func(serverAddr, httpAddr, openID string, messagesJSON string, repeatCount int, rangeStart, rangeEnd int, onProgress func(total, sent int, currentMsg string) bool) error {
		var messages []protocol.RecordMessage
		if err := json.Unmarshal([]byte(messagesJSON), &messages); err != nil {
			return fmt.Errorf("解析消息列表失败: %v", err)
		}
		// 使用 NewReplayer 复用连接池；ConnPool 会通过 PooledAuthenticator 自动包装
		return protocol.NewReplayer(protocol.ReplayOptions{
			ServerAddr:  serverAddr,
			HTTPAddr:    httpAddr,
			OpenID:      openID,
			Messages:    messages,
			RepeatCount: repeatCount,
			RangeStart:  rangeStart,
			RangeEnd:    rangeEnd,
			OnProgress:  onProgress,
			OnMessage: func(name string, msgID uint16, seqID uint32, payloadJSON string, offsetMs int, direction string, accountID string) {
				replayWorker.EmitReplayMessage(name, msgID, seqID, payloadJSON, offsetMs, direction, accountID)
			},
			ConnPool: connPool,
		}).SendMessages()
	})

	// 创建 proto-test 配置服务并自动启动监听
	protoTestConfigSvc := prototest.NewProtoTestConfigService()
	recordWorker := prototest.NewRecordWorker(app.Event)
	protoTestConfig, err := protoTestConfigSvc.GetConfig()
	if err != nil {
		log.Printf("[main] 读取 proto-test 配置失败: %v, 使用默认配置", err)
		protoTestConfig = prototest.DefaultProtoTestConfig()
	}
	if err := recordWorker.StartListen(protoTestConfig.TargetServerAddr, protoTestConfig.TargetHTTPAddr, protoTestConfig.TCPListenPort, protoTestConfig.HTTPListenPort); err != nil {
		log.Printf("[main] 启动 proto-test 监听失败: %v", err)
	}

	replayControlSvc := prototest.NewReplayControlService(replayWorker, connPool, recordWorker)
	recordControlSvc := prototest.NewRecordControlService(recordWorker, connPool)
	testCaseSvc := prototest.NewTestCaseService(recordFileSvc)
	app.RegisterService(application.NewService(recordFileSvc))
	app.RegisterService(application.NewService(replayControlSvc))
	app.RegisterService(application.NewService(recordControlSvc))
	app.RegisterService(application.NewService(testCaseSvc))
	app.RegisterService(application.NewService(protoTestConfigSvc))
	app.RegisterService(application.NewService(serverconfig.NewServerConfigService(excelConfigSvc)))

	// 注册聊天服务（注入业务模块服务，使聊天机器人可调用相应工具）
	app.RegisterService(application.NewService(home.NewChatServiceWithServices(app, &home.ChatServices{
		ExcelCheckService:    excelCheckService,
		RecordControlService: recordControlSvc,
		TestCaseService:      testCaseSvc,
		RecordFileService:    recordFileSvc,
	})))

	// 注册路线图服务
	app.RegisterService(application.NewService(roadmap.NewRoadmapService(app)))

	// 创建 MCP 配置服务（用于前端绑定）
	mcpConfigService := settings.NewMCPConfigService()
	app.RegisterService(application.NewService(mcpConfigService))

	// ========== 启动 MCP 服务器 ==========
	mcpConfig, err := mcp.LoadConfig()
	if err != nil {
		log.Printf("加载 MCP 配置失败: %v, 使用默认配置", err)
		mcpConfig = mcp.DefaultConfig()
	}

	mcpServices, err := mcp.BuildDefaultMcpServices()
	if err != nil {
		log.Printf("构建 HTTP MCP 服务失败: %v", err)
		return
	}
	// GUI 模式使用已经创建的 gameExcelSvc 实例，保持一致性
	mcpServices.McpGameExcelService = gameExcelSvc

	mcpServer := mcp.NewMCPServer(mcpConfig, mcpServices)

	mcpConfigService.SetServer(mcpServer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mcpServer.Start(ctx); err != nil {
		log.Printf("MCP server error: %v", err)
		return
	}
	defer mcpServer.Stop(ctx)

	// 创建主窗口
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Rain QA Func",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
		Width:            1600,
		MinWidth:         1200,
		Height:           900,
		MinHeight:        300,
		DevToolsEnabled:  true,
	})

	// 每秒发送时间事件到前端
	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	// 运行应用（阻塞直到退出）
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// excelDirAdapter 将 settings.ExcelConfigService 适配为 game.ExcelDirProvider 接口，
// 供 SkillUIDescService 读取统一策划配表目录，打破 game↔settings 循环依赖。
type excelDirAdapter struct {
	svc *settings.ExcelConfigService
}

func (a *excelDirAdapter) GetExcelDir() string {
	if cfg, err := a.svc.GetConfig(); err == nil && cfg != nil {
		return cfg.ExcelDir
	}
	return ""
}
