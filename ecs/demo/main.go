package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

// 全局日志函数，自动判空
func LogInfo(msg string) {
	if GlobalLogger != nil {
		GlobalLogger.logChan <- LogData{Message: msg, Level: "INFO", Time: time.Now()}
	}
}
func LogInfof(format string, args ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.logChan <- LogData{Message: fmt.Sprintf(format, args...), Level: "INFO", Time: time.Now()}
	}
}

func main() {
	logger := NewLoggerSystem("ecs.log")
	GlobalLogger = logger
	if GlobalLogger != nil {
		GlobalLogger.logChan <- LogData{Message: "main开始", Level: "INFO", Time: time.Now()}
	}

	// 启动CPU Profile
	f, err := os.Create("cpu.prof")
	if err == nil {
		_ = pprof.StartCPUProfile(f)
		defer func() {
			pprof.StopCPUProfile()
			f.Close()
		}()
	}

	var now = time.Now()
	ecs := NewECS()
	// 事件系统示例
	ecs.EventBus().Subscribe(EventKill, func(e Event) {
		LogInfo(fmt.Sprintf("[事件] Entity %d 被Kill, 原因: %v", e.Entity, e.Data))
	})
	// 高级事件系统示例
	ecs.AdvancedBus().Subscribe(EventGrowth, func(e Event) {
		LogInfo(fmt.Sprintf("[高级事件] Entity %d 成长: %v", e.Entity, e.Data))
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
	LogInfo("创建完成")

	// 动态增删component
	id := ecs.CreateEntity(map[ComponentType]interface{}{
		Attack:   AttackData{Value: 10},
		Velocity: VelocityData{X: 1, Y: 2},
	})
	LogInfo(fmt.Sprintf("Entity %d 初始: Attack+Velocity", id))
	ecs.AddComponent(id, Shape, ShapeData{Name: "Square"})
	LogInfo(fmt.Sprintf("Entity %d 增加Shape", id))
	ecs.RemoveComponent(id, Attack)
	LogInfo(fmt.Sprintf("Entity %d 移除Attack", id))

	// 系统扩展性与性能监控
	sm := NewSystemManager()
	hpSys := NewHPSystem()
	sm.Register(hpSys)
	sm.Register(NewAISystem())
	sm.Register(NewPhysicsSystem())
	sm.Register(NewRenderSystem())
	sm.SortSystems()

	// 并行处理所有system
	start := time.Now()
	sm.ParallelUpdateAll(ecs, 10*time.Millisecond)
	LogInfo(fmt.Sprintf("并行系统处理100万entity耗时: %v", time.Since(start)))

	// 复杂事件链测试
	ecs.AdvancedBus().Subscribe(EventHide, func(e Event) {
		LogInfo(fmt.Sprintf("[链式事件] Entity %d 被隐藏，自动触发成长", e.Entity))
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

	LogInfo(fmt.Sprintf("ECS系统复杂事件与system测试完成, 耗时: %v", time.Since(now)))

	// 生命周期钩子测试
	RegisterOnCreate(func(e *Entity) {
		LogInfo(fmt.Sprintf("[生命周期] 创建Entity %d", e.ID))
	})
	RegisterOnDestroy(func(e *Entity) {
		LogInfo(fmt.Sprintf("[生命周期] 销毁Entity %d", e.ID))
	})

	// 标签分组批量操作
	for i := 0; i < 10; i++ {
		_ = ecs.CreateEntityWithTags(map[ComponentType]interface{}{Attack: AttackData{Value: i}}, "enemy")
	}
	LogInfo(fmt.Sprintf("enemy标签下entity数量: %d", len(ecs.EntitiesWithTag("enemy"))))

	// 组件组合
	combo := CompositeComponent{Components: []ComponentType{Attack, Velocity}}
	LogInfo(fmt.Sprintf("组合组件bitmask: %d", combo.Type()))

	// System热插拔
	sm = NewSystemManager()
	sysA := NewHPSystem()
	sm.Register(sysA)
	sysB := NewAISystem()
	sm.HotSwap("HPSystem", sysB)
	LogInfo(fmt.Sprintf("System热插拔后第一个system: %s", sm.systems[0].Name()))

	// 快照/回滚
	snap := ecs.Snapshot()
	id = ecs.CreateEntityWithTags(map[ComponentType]interface{}{Attack: AttackData{Value: 99}}, "hero")
	LogInfo(fmt.Sprintf("创建新entity: %d", id))
	ecs.Restore(snap)
	LogInfo(fmt.Sprintf("回滚后hero标签下entity数量: %d", len(ecs.EntitiesWithTag("hero"))))

	// 事件优先级/过滤/重放
	peb := NewPriorityEventBus()
	peb.Subscribe(EventKill, EventSubscription{
		Priority: 10,
		Filter:   func(e Event) bool { return e.Entity%2 == 0 },
		Handler: func(e Event) {
			LogInfo(fmt.Sprintf("[高优先级偶数Kill] %d", e.Entity))
		},
	})
	peb.Subscribe(EventKill, EventSubscription{
		Priority: 1,
		Handler: func(e Event) {
			LogInfo(fmt.Sprintf("[低优先级Kill] %d", e.Entity))
		},
	})
	for i := 0; i < 5; i++ {
		peb.Publish(Event{Type: EventKill, Entity: EntityID(i)})
	}

	// 事件重放
	ebus := NewAdvancedEventBus()
	ebus.Publish(CustomEvent{Type: EventGrowth, Entity: 42, Data: "test"})
	log := ebus.SaveLog()
	LogInfo("事件日志保存，重放:")
	ebus.Replay(log)

	// 性能分析
	ecs.WriteProfile("ecs.prof")
	LogInfo("已写入heap profile: ecs.prof")

	// 脚本注册
	RegisterScriptSystem("PrintHello", func(args ...interface{}) {
		LogInfo(fmt.Sprintf("[脚本System] Hello %v", args))
	})
	scriptSystems["PrintHello"]("world", 123)

	// 配置加载
	cfg, err := LoadECSConfig("ecs_config.json")
	if err == nil {
		LogInfo(fmt.Sprintf("加载配置: %+v", cfg))
	} else {
		LogInfo("未找到配置文件，跳过配置加载测试")
	}

	// Mock
	mock := MockEntity(999, Attack|Velocity, "mock")
	LogInfo(fmt.Sprintf("Mock entity: %+v", mock))

	// 网络事件
	SendNetworkEvent(NetworkEvent{Type: EventKill, Entity: 888, Data: "net"})
	ReceiveNetworkEvent(NetworkEvent{Type: EventGrowth, Entity: 777, Data: "recv"})

	// 注册LoggerSystem
	sm.Register(logger)

	// 日志entity测试
	for i := 0; i < 5; i++ {
		_ = ecs.CreateEntity(map[ComponentType]interface{}{
			Log: LogData{Message: fmt.Sprintf("测试日志%d", i), Level: "INFO", Time: time.Now()},
		})
	}

	LogInfo("main中间")
	LogInfo(fmt.Sprintf("全部高级功能测试完成, 总耗时: %v", time.Since(now)))
	// 日志写入后sleep，保证异步日志全部落盘
	time.Sleep(2 * time.Second)
	close(logger.logChan)
	logger.file.Close()
}
