package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
	// 引入ECS主模块
)

func shouldProfile() bool {
	// 环境变量优先
	if os.Getenv("ECS_PROFILE") == "1" {
		return true
	}
	// 命令行参数
	for _, arg := range os.Args[1:] {
		if arg == "--profile" || arg == "-p" {
			return true
		}
	}
	return false
}

func main() {
	// 创建mainLogger Entity并挂载Log组件
	mainLoggerComps := map[ComponentType]interface{}{
		Log: LogData{Message: "main logger entity created", Level: "INFO", Time: time.Now()},
	}
	ecs := NewECS()
	mainLoggerEntityID := ecs.CreateEntity(mainLoggerComps)
	mainLogger := ecs.entities[mainLoggerEntityID]

	mainLogger.Log(ecs, "main通过Entity写入日志", "INFO")

	var cpuProfile *os.File
	if shouldProfile() {
		cpuProfile, _ = os.Create("cpu.prof")
		_ = pprof.StartCPUProfile(cpuProfile)
	}

	var now = time.Now()
	// 事件系统示例
	ecs.EventBus().Subscribe(EventKill, func(e Event) {
		mainLogger.Log(ecs, fmt.Sprintf("[事件] Entity %d 被Kill, 原因: %v", e.Entity, e.Data), "INFO")
	})
	// 高级事件系统示例
	ecs.AdvancedBus().Subscribe(EventGrowth, func(e Event) {
		mainLogger.Log(ecs, fmt.Sprintf("[高级事件] Entity %d 成长: %v", e.Entity, e.Data), "INFO")
	})

	// 创建100万个entity，分布不同component
	n := 1000000
	for i := 0; i < n; i++ {
		comps := make(map[ComponentType]interface{})
		if i%2 == 0 {
			comps[Attack] = AttackData{Value: rand.Intn(100)}
		}
		if i%3 == 0 {
			comps[Velocity] = VelocityData{X: rand.Float64() * 5, Y: rand.Float64() * 5}
		}
		if i%5 == 0 {
			comps[Shape] = ShapeData{Name: "Circle"}
		}
		if i%7 == 0 {
			comps[Color] = ColorData{R: rand.Intn(256), G: rand.Intn(256), B: rand.Intn(256)}
		}
		if i%11 == 0 {
			comps[Render] = RenderData{Mesh: "MeshA"}
		}
		if i%13 == 0 {
			comps[Kill] = KillData{Reason: "Test"}
		}
		if i%17 == 0 {
			comps[Hide] = HideData{Hidden: false}
		}
		if i%19 == 0 {
			comps[Growth] = GrowthData{Level: rand.Intn(10)}
		}
		_ = ecs.CreateEntity(comps)
	}
	mainLogger.Log(ecs, "创建完成", "INFO")

	// 动态增删component
	id := ecs.CreateEntity(map[ComponentType]interface{}{
		Attack:   AttackData{Value: 10},
		Velocity: VelocityData{X: 1, Y: 2},
	})
	mainLogger.Log(ecs, fmt.Sprintf("Entity %d 初始: Attack+Velocity", id), "INFO")
	ecs.AddComponent(id, Shape, ShapeData{Name: "Square"})
	mainLogger.Log(ecs, fmt.Sprintf("Entity %d 增加Shape", id), "INFO")
	ecs.RemoveComponent(id, Attack)
	mainLogger.Log(ecs, fmt.Sprintf("Entity %d 移除Attack", id), "INFO")

	// 系统扩展性与性能监控
	sm := NewSystemManager()
	hpSys := NewHPSystem()
	sm.Register(hpSys)
	sm.Register(NewAISystem())
	sm.Register(NewPhysicsSystem())
	sm.Register(NewRenderSystem())
	sm.SortSystems()

	// 注册LoggerSystem
	logger := NewLoggerSystem("demo/ecs.log")
	sm.Register(logger)

	// 并行处理所有system
	start := time.Now()
	sm.ParallelUpdateAll(ecs, 10*time.Millisecond)
	mainLogger.Log(ecs, fmt.Sprintf("并行系统处理100万entity耗时: %v", time.Since(start)), "INFO")

	// 复杂事件链测试
	ecs.AdvancedBus().Subscribe(EventHide, func(e Event) {
		mainLogger.Log(ecs, fmt.Sprintf("[链式事件] Entity %d 被隐藏，自动触发成长", e.Entity), "INFO")
		ecs.AdvancedBus().Publish(CustomEvent{
			Type:   EventGrowth,
			Entity: e.Entity,
			Data:   "链式成长",
		})
	})
	ecs.AdvancedBus().Publish(CustomEvent{
		Type:   EventHide,
		Entity: 42,
		Data:   "链式隐藏",
	})

	// 性能监控
	sm.PrintStats()

	// 批量事件：让所有entity沉睡
	ids := make([]EntityID, 0, 1000)
	for i := 0; i < 1000; i++ {
		ids = append(ids, EntityID(i))
	}
	ecs.AdvancedBus().Publish(CustomEvent{
		Type:    EventHide,
		Targets: ids,
		Data:    "全体沉睡",
	})

	// 延迟事件：3秒后全体复活
	ecs.AdvancedBus().Publish(CustomEvent{
		Type:    EventGrowth,
		Targets: ids,
		Data:    "全体复活",
		Delay:   3 * time.Second,
	})

	// 区域事件：假设Area=1的entity批量成长
	areaIDs := make([]EntityID, 0)
	for i := 0; i < 1000; i++ {
		if i%10 == 0 {
			areaIDs = append(areaIDs, EntityID(i))
		}
	}
	ecs.AdvancedBus().Publish(CustomEvent{
		Type:    EventGrowth,
		Targets: areaIDs,
		Data:    "区域成长",
		Area:    1,
	})

	// Tick高级事件系统，模拟帧循环
	for i := 0; i < 400; i++ {
		ecs.AdvancedBus().Tick()
		time.Sleep(10 * time.Millisecond)
	}

	mainLogger.Log(ecs, fmt.Sprintf("ECS系统复杂事件与system测试完成, 耗时: %v", time.Since(now)), "INFO")

	// 生命周期钩子测试
	RegisterOnCreate(func(e *Entity) {
		mainLogger.Log(ecs, fmt.Sprintf("[生命周期] 创建Entity %d", e.ID), "INFO")
	})
	RegisterOnDestroy(func(e *Entity) {
		mainLogger.Log(ecs, fmt.Sprintf("[生命周期] 销毁Entity %d", e.ID), "INFO")
	})

	// 标签分组批量操作
	for i := 0; i < 10; i++ {
		_ = ecs.CreateEntityWithTags(map[ComponentType]interface{}{Attack: AttackData{Value: i}}, "enemy")
	}
	mainLogger.Log(ecs, fmt.Sprintf("enemy标签下entity数量: %d", len(ecs.EntitiesWithTag("enemy"))), "INFO")

	// 组件组合
	combo := CompositeComponent{Components: []ComponentType{Attack, Velocity}}
	mainLogger.Log(ecs, fmt.Sprintf("组合组件bitmask: %d", combo.Type()), "INFO")

	// System热插拔
	sm = NewSystemManager()
	sysA := NewHPSystem()
	sm.Register(sysA)
	sysB := NewAISystem()
	sm.HotSwap("HPSystem", sysB)
	mainLogger.Log(ecs, fmt.Sprintf("System热插拔后第一个system: %s", sm.systems[0].Name()), "INFO")

	// 快照/回滚
	snap := ecs.Snapshot()
	id = ecs.CreateEntityWithTags(map[ComponentType]interface{}{Attack: AttackData{Value: 99}}, "hero")
	mainLogger.Log(ecs, fmt.Sprintf("创建新entity: %d", id), "INFO")
	ecs.Restore(snap)
	mainLogger.Log(ecs, fmt.Sprintf("回滚后hero标签下entity数量: %d", len(ecs.EntitiesWithTag("hero"))), "INFO")

	// 事件优先级/过滤/重放
	peb := NewPriorityEventBus()
	peb.Subscribe(EventKill, EventSubscription{
		Priority: 10,
		Filter:   func(e Event) bool { return e.Entity%2 == 0 },
		Handler: func(e Event) {
			mainLogger.Log(ecs, fmt.Sprintf("[高优先级偶数Kill] %d", e.Entity), "INFO")
		},
	})
	peb.Subscribe(EventKill, EventSubscription{
		Priority: 1,
		Handler: func(e Event) {
			mainLogger.Log(ecs, fmt.Sprintf("[低优先级Kill] %d", e.Entity), "INFO")
		},
	})
	for i := 0; i < 5; i++ {
		peb.Publish(Event{Type: EventKill, Entity: EntityID(i)})
	}

	// 事件重放
	ebus := NewAdvancedEventBus()
	ebus.Publish(CustomEvent{Type: EventGrowth, Entity: 42, Data: "test"})
	log := ebus.SaveLog()
	mainLogger.Log(ecs, "事件日志保存，重放:", "INFO")
	ebus.Replay(log)

	// 性能分析
	if shouldProfile() {
		ecs.WriteProfile("ecs.prof")
		mainLogger.Log(ecs, "已写入heap profile: ecs.prof", "INFO")
	}

	// 脚本注册
	RegisterScriptSystem("PrintHello", func(args ...interface{}) {
		mainLogger.Log(ecs, fmt.Sprintf("[脚本System] Hello %v", args), "INFO")
	})
	scriptSystems["PrintHello"]("world", 123)

	// 配置加载
	cfg, err := LoadECSConfig("ecs_config.json")
	if err == nil {
		mainLogger.Log(ecs, fmt.Sprintf("加载配置: %+v", cfg), "INFO")
	} else {
		mainLogger.Log(ecs, "未找到配置文件，跳过配置加载测试", "INFO")
	}

	// Mock
	mock := MockEntity(999, Attack|Velocity, "mock")
	mainLogger.Log(ecs, fmt.Sprintf("Mock entity: %+v", mock), "INFO")

	// 网络事件
	SendNetworkEvent(NetworkEvent{Type: EventKill, Entity: 888, Data: "net"})
	ReceiveNetworkEvent(NetworkEvent{Type: EventGrowth, Entity: 777, Data: "recv"})

	// 日志entity测试
	for i := 0; i < 5; i++ {
		_ = ecs.CreateEntity(map[ComponentType]interface{}{
			Log: LogData{Message: fmt.Sprintf("测试日志%d", i), Level: "INFO", Time: time.Now()},
		})
	}

	mainLogger.Log(ecs, "main中间", "INFO")
	mainLogger.Log(ecs, fmt.Sprintf("全部高级功能测试完成, 总耗时: %v", time.Since(now)), "INFO")
	// 日志写入后，强制更新所有system，确保日志entity被LoggerSystem消费
	sm.UpdateAll(ecs, 0)
	// 手动调用一次logger.Update(ecs)调试
	logger.Update(ecs)
	// 日志写入后sleep，保证异步日志全部落盘
	time.Sleep(2 * time.Second)
	close(logger.logChan)
	logger.file.Close()

	if shouldProfile() && cpuProfile != nil {
		pprof.StopCPUProfile()
		cpuProfile.Close()
	}
}
