package main

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

// 全局LoggerSystem，便于系统/钩子/事件输出日志
var GlobalLogger *LoggerSystem

type SystemPriority int

type SystemEx interface {
	System
	Priority() SystemPriority
	DependsOn() []string
	Name() string
	Filter(e *Entity, ecs *ECS) bool
	Tick(ecs *ECS, dt time.Duration)
}

type BaseSystem struct {
	name     string
	priority SystemPriority
	depends  []string
	filter   func(*Entity, *ECS) bool
	lastTick time.Time
	interval time.Duration
	calls    int
	cost     time.Duration
	active   bool
}

func (s *BaseSystem) Priority() SystemPriority { return s.priority }
func (s *BaseSystem) DependsOn() []string      { return s.depends }
func (s *BaseSystem) Name() string             { return s.name }
func (s *BaseSystem) Filter(e *Entity, ecs *ECS) bool {
	if s.filter != nil {
		return s.filter(e, ecs)
	}
	return true
}
func (s *BaseSystem) Tick(ecs *ECS, dt time.Duration) {}

// 示例：定时执行、条件过滤、性能监控

type HPData struct{ Value int }
type BuffData struct {
	Name     string
	Duration int
}
type AreaData struct{ ID int }

type HPSystem struct{ BaseSystem }

func NewHPSystem() *HPSystem {
	return &HPSystem{BaseSystem{
		name:     "HPSystem",
		priority: 1,
		filter: func(e *Entity, ecs *ECS) bool {
			return e.arch&Attack != 0 // 只处理有Attack的entity
		},
		interval: 100 * time.Millisecond,
		active:   true,
	}}
}

func (s *HPSystem) Update(ecs *ECS) {
	start := time.Now()
	for _, ent := range ecs.entities {
		if !s.Filter(ent, ecs) {
			continue
		}
		// 假设HP随Attack变化
		chunk := ecs.archMgr.GetOrCreateChunk(ent.arch)
		if ent.arch&Attack != 0 && len(chunk.attack) > ent.chunkIdx {
			val := chunk.attack[ent.chunkIdx].Value
			if val > 50 {
				// 触发高攻击事件
				if ecs.eventBus != nil {
					ecs.eventBus.Publish(Event{Type: EventGrowth, Entity: ent.ID, Data: "高攻击"})
				}
			}
		}
	}
	s.cost += time.Since(start)
	s.calls++
}

// 定时执行
func (s *HPSystem) Tick(ecs *ECS, dt time.Duration) {
	if time.Since(s.lastTick) >= s.interval {
		s.Update(ecs)
		s.lastTick = time.Now()
	}
}

// 动态注册/注销system

type SystemManager struct {
	systems []SystemEx
}

func NewSystemManager() *SystemManager {
	return &SystemManager{systems: make([]SystemEx, 0)}
}

func (sm *SystemManager) Register(sys SystemEx) {
	sm.systems = append(sm.systems, sys)
}

func (sm *SystemManager) Unregister(name string) {
	for i, s := range sm.systems {
		if s.Name() == name {
			sm.systems = append(sm.systems[:i], sm.systems[i+1:]...)
			break
		}
	}
}

func (sm *SystemManager) UpdateAll(ecs *ECS, dt time.Duration) {
	for _, s := range sm.systems {
		if sys, ok := s.(interface{ Tick(*ECS, time.Duration) }); ok {
			sys.Tick(ecs, dt)
		} else {
			s.Update(ecs)
		}
	}
}

func (sm *SystemManager) PrintStats() {
	for _, s := range sm.systems {
		switch sys := s.(type) {
		case *HPSystem:
			if GlobalLogger != nil {
				GlobalLogger.logChan <- LogData{
					Message: fmt.Sprintf("System %s 调用%d次, 总耗时%v", sys.Name(), sys.calls, sys.cost),
					Level:   "INFO",
					Time:    time.Now(),
				}
			}
			// 可扩展更多system类型
		}
	}
}

// 并行执行所有system
func (sm *SystemManager) ParallelUpdateAll(ecs *ECS, dt time.Duration) {
	var wg sync.WaitGroup
	for _, s := range sm.systems {
		wg.Add(1)
		go func(sys SystemEx) {
			defer wg.Done()
			if t, ok := sys.(interface{ Tick(*ECS, time.Duration) }); ok {
				t.Tick(ecs, dt)
			} else {
				sys.Update(ecs)
			}
		}(s)
	}
	wg.Wait()
}

// 按优先级排序（可扩展为拓扑排序实现依赖）
func (sm *SystemManager) SortSystems() {
	sort.Slice(sm.systems, func(i, j int) bool {
		return sm.systems[i].Priority() > sm.systems[j].Priority()
	})
}

// AI System

type AISystem struct{ BaseSystem }

func NewAISystem() *AISystem {
	return &AISystem{BaseSystem{
		name:     "AISystem",
		priority: 2,
		interval: 50 * time.Millisecond,
		active:   true,
	}}
}
func (s *AISystem) Update(ecs *ECS) {
	for _, ent := range ecs.entities {
		if ent.arch&Velocity != 0 {
			chunk := ecs.archMgr.GetOrCreateChunk(ent.arch)
			if len(chunk.velocity) > ent.chunkIdx {
				chunk.velocity[ent.chunkIdx].X += 0.1
				chunk.velocity[ent.chunkIdx].Y += 0.1
			}
		}
	}
}

// 物理System

type PhysicsSystem struct{ BaseSystem }

func NewPhysicsSystem() *PhysicsSystem {
	return &PhysicsSystem{BaseSystem{
		name:     "PhysicsSystem",
		priority: 3,
		interval: 20 * time.Millisecond,
		active:   true,
	}}
}
func (s *PhysicsSystem) Update(ecs *ECS) {
	// 物理逻辑：如碰撞检测、位置更新等
}

// 渲染System

type RenderSystem struct{ BaseSystem }

func NewRenderSystem() *RenderSystem {
	return &RenderSystem{BaseSystem{
		name:     "RenderSystem",
		priority: 0,
		interval: 16 * time.Millisecond,
		active:   true,
	}}
}
func (s *RenderSystem) Update(ecs *ECS) {
	// 渲染逻辑：如输出entity状态
}

// System热插拔/快照/回滚
func (sm *SystemManager) HotSwap(oldName string, newSys SystemEx) {
	for i, s := range sm.systems {
		if s.Name() == oldName {
			sm.systems[i] = newSys
			return
		}
	}
	sm.Register(newSys)
}

// ECS快照/回滚（深拷贝，简化版）
func (ecs *ECS) Snapshot() *ECS {
	ecs.lock.RLock()
	defer ecs.lock.RUnlock()
	copyECS := NewECS()
	for id, ent := range ecs.entities {
		newEnt := *ent
		copyECS.entities[id] = &newEnt
	}
	// chunk和component数据可按需深拷贝
	return copyECS
}

func (ecs *ECS) Restore(snap *ECS) {
	ecs.lock.Lock()
	defer ecs.lock.Unlock()
	ecs.entities = snap.entities
	// chunk和component数据可按需恢复
}

// LoggerSystem: 异步写入日志到文件

type LoggerSystem struct {
	logChan chan LogData
	file    *os.File
}

func NewLoggerSystem(filename string) *LoggerSystem {
	f, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	sys := &LoggerSystem{
		logChan: make(chan LogData, 1000),
		file:    f,
	}
	go sys.run()
	return sys
}

func (s *LoggerSystem) run() {
	count := 0
	for log := range s.logChan {
		fmt.Fprintf(s.file, "[%s] %s: %s\n", log.Time.Format(time.RFC3339), log.Level, log.Message)
		count++
		if count%100 == 0 {
			s.file.Sync()
		}
	}
	s.file.Sync() // 关闭前强制落盘
}

func (s *LoggerSystem) Update(ecs *ECS) {
	for arch, chunk := range ecs.archMgr.chunks {
		if chunk.archetype&Log != 0 {
			msg := fmt.Sprintf("[LoggerSystem] chunk arch=0x%x entityCount=%d", arch, len(chunk.entities))
			s.logChan <- LogData{Message: msg, Level: "DEBUG", Time: time.Now()}
			for i, logs := range chunk.log {
				msg := fmt.Sprintf("  entityIdx=%d entityID=%v logCount=%d", i, chunk.entities[i], len(logs))
				s.logChan <- LogData{Message: msg, Level: "DEBUG", Time: time.Now()}
				for _, log := range logs {
					s.logChan <- log
				}
				chunk.log[i] = nil // 消费后清空队列
			}
		}
	}
}

func (s *LoggerSystem) Priority() SystemPriority        { return 0 }
func (s *LoggerSystem) DependsOn() []string             { return nil }
func (s *LoggerSystem) Name() string                    { return "LoggerSystem" }
func (s *LoggerSystem) Filter(e *Entity, ecs *ECS) bool { return true }
func (s *LoggerSystem) Tick(ecs *ECS, dt time.Duration) {}
