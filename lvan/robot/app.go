package main

import "syscall"

var defaultApp = newApp()

func Init() error { return defaultApp.init() }

func Run() error {
	return defaultApp.run()
}

func Stop() {
	defaultApp.cancel()
}

type app struct {
	modules module.Modules
	ctx context.Context
	cancel context.CancelFunc
}

func newApp() *app {
	ctx, cancel := context.WithCancel(context.Background())
	return &app{
		ctx: ctx,
		cancel: cancel,
	}
}
func (a *app) init() error {
	if err := clogger.Init(cConfig.Robot); err != nil {
		logger.Error(msg: "logger init failed", args...:"error", err)
		return err
	}

	if err := config.InitConfig(); err != nil {
		logger.Error(msg: "init config failed", args...:"error", err)
		return err
	}

	// 加载配置后重新初始化log
	if err := clogger.Init(cConfig.Robot); err != nil {
		logger.Error(msg: "config logger failed", args...:"error", err)
		return err
	}

	// 注册信号处理函数
	stop := func() error {
		logger.Warn(msg: "stop robots start")
		Stop()
		return nil
	}

	signal.RegisterProc(syscall.SIGABRT, stop)
	signal.RegisterProc(syscall.SIGTERM, stop)
	signal.RegisterProc(syscall.SIGINT, stop)

	robot.NewRobots(a.ctx)

	if err := a.modules.Register(
		robot.GetModule());
		err != nil { return err }

	return nil
}
func (a *app) run() error { // 1 usage   张锦
	if err := a.modules.Init(); err != nil {
		return err
	}

	if err := a.modules.Start(); err != nil {
		return err
	}

	a.modules.Stop()
	a.modules.Close()

	return nil
}

func (a *app) Stop() {  // 张锦
	a.modules.Stop()
}

func (a *app) Close() {  // 张锦
	a.modules.Close()
}
