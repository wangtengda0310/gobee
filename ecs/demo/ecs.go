package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"sync"
	"time"
)

// ComponentType 用bitmask表示
// 这里只支持最多64种component
// 真实项目可用map扩展

type ComponentType uint64

const (
	Attack ComponentType = 1 << iota
	Velocity
	Shape
	Color
	Render
	Kill
	Hide
	Growth
	Log
)

type EntityID int

type EventType int

const (
	EventKill EventType = iota
	EventHide
	EventGrowth
)

type Event struct {
	Type   EventType
	Entity EntityID
	Data   interface{}
}

type EventListener func(Event)

type EventBus struct {
	listeners map[EventType][]EventListener
	lock      sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{listeners: make(map[EventType][]EventListener)}
}

func (bus *EventBus) Subscribe(eventType EventType, listener EventListener) {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	bus.listeners[eventType] = append(bus.listeners[eventType], listener)
}

func (bus *EventBus) Publish(event Event) {
	bus.lock.RLock()
	defer bus.lock.RUnlock()
	for _, l := range bus.listeners[event.Type] {
		l(event)
	}
}

// Component数据定义
// 不同类型component用不同slice存储

type AttackData struct{ Value int }
type VelocityData struct{ X, Y float64 }
type ShapeData struct{ Name string }
type ColorData struct{ R, G, B int }
type RenderData struct{ Mesh string }
type KillData struct{ Reason string }
type HideData struct{ Hidden bool }
type GrowthData struct{ Level int }
type LogData struct {
	Message string
	Level   string
	Time    time.Time
}

// Chunk: SoA结构

type Chunk struct {
	archetype ComponentType // bitmask
	entities  []EntityID
	attack    []AttackData
	velocity  []VelocityData
	shape     []ShapeData
	color     []ColorData
	render    []RenderData
	kill      []KillData
	hide      []HideData
	growth    []GrowthData
	log       [][]LogData // 每个entity一个日志队列
}

// 全局chunk池
var chunkPool = sync.Pool{
	New: func() interface{} { return &Chunk{} },
}

func GetChunk() *Chunk {
	return chunkPool.Get().(*Chunk)
}

func PutChunk(c *Chunk) {
	c.entities = c.entities[:0]
	c.attack = c.attack[:0]
	c.velocity = c.velocity[:0]
	c.shape = c.shape[:0]
	c.color = c.color[:0]
	c.render = c.render[:0]
	c.kill = c.kill[:0]
	c.hide = c.hide[:0]
	c.growth = c.growth[:0]
	c.log = c.log[:0]
	chunkPool.Put(c)
}

func (c *Chunk) AddEntity(e EntityID, comps map[ComponentType]interface{}) {
	c.entities = append(c.entities, e)
	if c.archetype&Attack != 0 {
		c.attack = append(c.attack, comps[Attack].(AttackData))
	}
	if c.archetype&Velocity != 0 {
		c.velocity = append(c.velocity, comps[Velocity].(VelocityData))
	}
	if c.archetype&Shape != 0 {
		c.shape = append(c.shape, comps[Shape].(ShapeData))
	}
	if c.archetype&Color != 0 {
		c.color = append(c.color, comps[Color].(ColorData))
	}
	if c.archetype&Render != 0 {
		c.render = append(c.render, comps[Render].(RenderData))
	}
	if c.archetype&Kill != 0 {
		c.kill = append(c.kill, comps[Kill].(KillData))
	}
	if c.archetype&Hide != 0 {
		c.hide = append(c.hide, comps[Hide].(HideData))
	}
	if c.archetype&Growth != 0 {
		c.growth = append(c.growth, comps[Growth].(GrowthData))
	}
	if c.archetype&Log != 0 {
		c.log = append(c.log, []LogData{comps[Log].(LogData)})
	}
}

func (c *Chunk) RemoveEntity(idx int) {
	c.entities = append(c.entities[:idx], c.entities[idx+1:]...)
	if c.archetype&Attack != 0 {
		c.attack = append(c.attack[:idx], c.attack[idx+1:]...)
	}
	if c.archetype&Velocity != 0 {
		c.velocity = append(c.velocity[:idx], c.velocity[idx+1:]...)
	}
	if c.archetype&Shape != 0 {
		c.shape = append(c.shape[:idx], c.shape[idx+1:]...)
	}
	if c.archetype&Color != 0 {
		c.color = append(c.color[:idx], c.color[idx+1:]...)
	}
	if c.archetype&Render != 0 {
		c.render = append(c.render[:idx], c.render[idx+1:]...)
	}
	if c.archetype&Kill != 0 {
		c.kill = append(c.kill[:idx], c.kill[idx+1:]...)
	}
	if c.archetype&Hide != 0 {
		c.hide = append(c.hide[:idx], c.hide[idx+1:]...)
	}
	if c.archetype&Growth != 0 {
		c.growth = append(c.growth[:idx], c.growth[idx+1:]...)
	}
	if c.archetype&Log != 0 {
		c.log = append(c.log[:idx], c.log[idx+1:]...)
	}
}

// Archetype管理

type ArchetypeManager struct {
	chunks map[ComponentType]*Chunk
	lock   sync.RWMutex
}

func NewArchetypeManager() *ArchetypeManager {
	return &ArchetypeManager{chunks: make(map[ComponentType]*Chunk)}
}

func (am *ArchetypeManager) GetOrCreateChunk(arch ComponentType) *Chunk {
	am.lock.Lock()
	defer am.lock.Unlock()
	chunk, ok := am.chunks[arch]
	if !ok {
		chunk = GetChunk()
		chunk.archetype = arch
		am.chunks[arch] = chunk
	}
	return chunk
}

func (am *ArchetypeManager) RemoveEntityFromChunk(arch ComponentType, idx int) {
	am.lock.Lock()
	defer am.lock.Unlock()
	if chunk, ok := am.chunks[arch]; ok {
		chunk.RemoveEntity(idx)
		if len(chunk.entities) == 0 {
			delete(am.chunks, arch)
			PutChunk(chunk)
		}
	}
}

// Entity管理

type Tag string

type LifecycleHook func(e *Entity)

var onCreateHooks []LifecycleHook
var onDestroyHooks []LifecycleHook

func RegisterOnCreate(h LifecycleHook)  { onCreateHooks = append(onCreateHooks, h) }
func RegisterOnDestroy(h LifecycleHook) { onDestroyHooks = append(onDestroyHooks, h) }

type Entity struct {
	ID       EntityID
	arch     ComponentType
	chunkIdx int // 在chunk中的下标
	Tags     []Tag
}

// 日志写入方法：只有挂载Log组件的Entity才可用
func (e *Entity) Log(ecs *ECS, msg string, level string) {
	if e.arch&Log == 0 {
		return // 未挂载Log组件
	}
	chunk := ecs.archMgr.GetOrCreateChunk(e.arch)
	if len(chunk.log) > e.chunkIdx {
		chunk.log[e.chunkIdx] = append(chunk.log[e.chunkIdx], LogData{
			Message: msg,
			Level:   level,
			Time:    time.Now(),
		})
	}
}

type ECS struct {
	entities   map[EntityID]*Entity
	archMgr    *ArchetypeManager
	eventBus   *EventBus
	advBus     *AdvancedEventBus
	nextEntity EntityID
	lock       sync.RWMutex
	groups     map[Tag][]EntityID
}

func NewECS() *ECS {
	return &ECS{
		entities: make(map[EntityID]*Entity),
		archMgr:  NewArchetypeManager(),
		eventBus: NewEventBus(),
		advBus:   NewAdvancedEventBus(),
		groups:   make(map[Tag][]EntityID),
	}
}

func (ecs *ECS) CreateEntityWithTags(comps map[ComponentType]interface{}, tags ...Tag) EntityID {
	ecs.lock.Lock()
	defer ecs.lock.Unlock()
	var arch ComponentType
	for t := range comps {
		arch |= t
	}
	chunk := ecs.archMgr.GetOrCreateChunk(arch)
	id := ecs.nextEntity
	chunk.AddEntity(id, comps)
	ent := &Entity{ID: id, arch: arch, chunkIdx: len(chunk.entities) - 1, Tags: tags}
	ecs.entities[id] = ent
	for _, tag := range tags {
		ecs.groups[tag] = append(ecs.groups[tag], id)
	}
	ecs.nextEntity++
	for _, h := range onCreateHooks {
		h(ent)
	}
	return id
}

func (ecs *ECS) DestroyEntity(id EntityID) {
	ecs.lock.Lock()
	defer ecs.lock.Unlock()
	ent, ok := ecs.entities[id]
	if !ok {
		return
	}
	for _, h := range onDestroyHooks {
		h(ent)
	}
	// 从分组移除
	for _, tag := range ent.Tags {
		ids := ecs.groups[tag]
		for i, eid := range ids {
			if eid == id {
				ecs.groups[tag] = append(ids[:i], ids[i+1:]...)
				break
			}
		}
	}
	delete(ecs.entities, id)
}

func (ecs *ECS) EntitiesWithTag(tag Tag) []*Entity {
	ecs.lock.RLock()
	defer ecs.lock.RUnlock()
	ids := ecs.groups[tag]
	result := make([]*Entity, 0, len(ids))
	for _, id := range ids {
		if ent, ok := ecs.entities[id]; ok {
			result = append(result, ent)
		}
	}
	return result
}

func (ecs *ECS) CreateEntity(comps map[ComponentType]interface{}) EntityID {
	ecs.lock.Lock()
	defer ecs.lock.Unlock()
	var arch ComponentType
	for t := range comps {
		arch |= t
	}
	chunk := ecs.archMgr.GetOrCreateChunk(arch)
	id := ecs.nextEntity
	chunk.AddEntity(id, comps)
	ent := &Entity{ID: id, arch: arch, chunkIdx: len(chunk.entities) - 1}
	ecs.entities[id] = ent
	ecs.nextEntity++
	return id
}

func (ecs *ECS) AddComponent(id EntityID, t ComponentType, data interface{}) {
	ecs.lock.Lock()
	defer ecs.lock.Unlock()
	ent := ecs.entities[id]
	oldArch := ent.arch
	newArch := oldArch | t
	if oldArch == newArch {
		return // 已有该组件
	}
	// 迁移到新chunk
	oldChunk := ecs.archMgr.GetOrCreateChunk(oldArch)
	newChunk := ecs.archMgr.GetOrCreateChunk(newArch)
	// 收集旧数据
	comps := make(map[ComponentType]interface{})
	if oldArch&Attack != 0 {
		comps[Attack] = oldChunk.attack[ent.chunkIdx]
	}
	if oldArch&Velocity != 0 {
		comps[Velocity] = oldChunk.velocity[ent.chunkIdx]
	}
	if oldArch&Shape != 0 {
		comps[Shape] = oldChunk.shape[ent.chunkIdx]
	}
	if oldArch&Color != 0 {
		comps[Color] = oldChunk.color[ent.chunkIdx]
	}
	if oldArch&Render != 0 {
		comps[Render] = oldChunk.render[ent.chunkIdx]
	}
	if oldArch&Kill != 0 {
		comps[Kill] = oldChunk.kill[ent.chunkIdx]
	}
	if oldArch&Hide != 0 {
		comps[Hide] = oldChunk.hide[ent.chunkIdx]
	}
	if oldArch&Growth != 0 {
		comps[Growth] = oldChunk.growth[ent.chunkIdx]
	}
	// 新增组件
	comps[t] = data
	// 添加到新chunk
	newChunk.AddEntity(id, comps)
	// 从旧chunk移除
	oldChunk.RemoveEntity(ent.chunkIdx)
	// 更新entity
	ent.arch = newArch
	ent.chunkIdx = len(newChunk.entities) - 1
}

func (ecs *ECS) RemoveComponent(id EntityID, t ComponentType) {
	ecs.lock.Lock()
	defer ecs.lock.Unlock()
	ent := ecs.entities[id]
	oldArch := ent.arch
	newArch := oldArch &^ t
	if oldArch == newArch {
		return // 没有该组件
	}
	oldChunk := ecs.archMgr.GetOrCreateChunk(oldArch)
	newChunk := ecs.archMgr.GetOrCreateChunk(newArch)
	comps := make(map[ComponentType]interface{})
	if newArch&Attack != 0 {
		comps[Attack] = oldChunk.attack[ent.chunkIdx]
	}
	if newArch&Velocity != 0 {
		comps[Velocity] = oldChunk.velocity[ent.chunkIdx]
	}
	if newArch&Shape != 0 {
		comps[Shape] = oldChunk.shape[ent.chunkIdx]
	}
	if newArch&Color != 0 {
		comps[Color] = oldChunk.color[ent.chunkIdx]
	}
	if newArch&Render != 0 {
		comps[Render] = oldChunk.render[ent.chunkIdx]
	}
	if newArch&Kill != 0 {
		comps[Kill] = oldChunk.kill[ent.chunkIdx]
	}
	if newArch&Hide != 0 {
		comps[Hide] = oldChunk.hide[ent.chunkIdx]
	}
	if newArch&Growth != 0 {
		comps[Growth] = oldChunk.growth[ent.chunkIdx]
	}
	newChunk.AddEntity(id, comps)
	oldChunk.RemoveEntity(ent.chunkIdx)
	ent.arch = newArch
	ent.chunkIdx = len(newChunk.entities) - 1
}

// System接口

type System interface {
	Update(ecs *ECS)
}

// 示例System

type VelocitySystem struct{}

func (s *VelocitySystem) Update(ecs *ECS) {
	jobs := []func(){}
	ecs.archMgr.lock.RLock()
	for _, chunk := range ecs.archMgr.chunks {
		if chunk.archetype&Velocity != 0 {
			c := chunk
			jobs = append(jobs, func() {
				vs := c.velocity
				n := len(vs)
				for i := 0; i < n; i += 4 {
					if i < n {
						vs[i].X += 1
						vs[i].Y += 1
					}
					if i+1 < n {
						vs[i+1].X += 1
						vs[i+1].Y += 1
					}
					if i+2 < n {
						vs[i+2].X += 1
						vs[i+2].Y += 1
					}
					if i+3 < n {
						vs[i+3].X += 1
						vs[i+3].Y += 1
					}
				}
			})
		}
	}
	ecs.archMgr.lock.RUnlock()
	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Add(1)
		go func(j func()) { defer wg.Done(); j() }(job)
	}
	wg.Wait()
}

type AttackSystem struct{}

func (s *AttackSystem) Update(ecs *ECS) {
	chunk := ecs.archMgr.GetOrCreateChunk(Attack)
	for i := range chunk.attack {
		chunk.attack[i].Value += 1
	}
}

// 更多System可扩展...

// 测试用例和主流程见 demo/main.go

func (ecs *ECS) EventBus() *EventBus {
	return ecs.eventBus
}

func (ecs *ECS) AdvancedBus() *AdvancedEventBus {
	return ecs.advBus
}

// Component继承/组合/Tag

type TagComponent struct{}

func (TagComponent) Type() ComponentType { return 0 } // 0代表Tag类型

type CompositeComponent struct {
	Components []ComponentType
}

func (c CompositeComponent) Type() ComponentType {
	var t ComponentType
	for _, ct := range c.Components {
		t |= ct
	}
	return t
}

func (ecs *ECS) WriteProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return pprof.WriteHeapProfile(f)
}

// 脚本与热更新接口（伪代码/预留）
type ScriptFunc func(args ...interface{})

var scriptSystems = map[string]ScriptFunc{}
var scriptEvents = map[string]ScriptFunc{}

func RegisterScriptSystem(name string, fn ScriptFunc) {
	scriptSystems[name] = fn
}
func RegisterScriptEvent(name string, fn ScriptFunc) {
	scriptEvents[name] = fn
}

type ECSConfig struct {
	Archetypes []map[string]interface{} `json:"archetypes"`
	Systems    []string                 `json:"systems"`
	Events     []string                 `json:"events"`
}

func LoadECSConfig(filename string) (*ECSConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cfg ECSConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// 单元测试与Mock能力
// 伪代码示例
func MockEntity(id EntityID, arch ComponentType, tags ...Tag) *Entity {
	return &Entity{ID: id, arch: arch, chunkIdx: 0, Tags: tags}
}

// 分布式同步与网络事件接口（伪代码/预留）
type NetworkEvent struct {
	Type   EventType
	Entity EntityID
	Data   interface{}
}

func SendNetworkEvent(e NetworkEvent) {
	// 伪代码：通过gRPC/WebSocket发送事件
}

func ReceiveNetworkEvent(e NetworkEvent) {
	// 伪代码：接收并分发事件
}
