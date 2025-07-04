package main

import (
	"sync"
	"time"
)

type CustomEvent struct {
	Type      EventType
	Entity    EntityID
	Data      interface{}
	Delay     time.Duration // 延迟触发
	Targets   []EntityID    // 批量事件
	Area      int           // 区域事件
	Triggered bool          // 是否已触发
}

type AdvancedEventBus struct {
	listeners map[EventType][]EventListener
	queue     []CustomEvent
	lock      sync.RWMutex
}

func NewAdvancedEventBus() *AdvancedEventBus {
	return &AdvancedEventBus{
		listeners: make(map[EventType][]EventListener),
		queue:     make([]CustomEvent, 0),
	}
}

func (bus *AdvancedEventBus) Subscribe(eventType EventType, listener EventListener) {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	bus.listeners[eventType] = append(bus.listeners[eventType], listener)
}

// 支持立即、延迟、批量、区域事件
func (bus *AdvancedEventBus) Publish(event CustomEvent) {
	bus.lock.Lock()
	if event.Delay > 0 {
		bus.queue = append(bus.queue, event)
		bus.lock.Unlock()
		return
	}
	listeners := append([]EventListener{}, bus.listeners[event.Type]...) // 拷贝一份
	bus.lock.Unlock()
	bus.dispatchWithListeners(event, listeners)
}

func (bus *AdvancedEventBus) dispatchWithListeners(event CustomEvent, listeners []EventListener) {
	for _, l := range listeners {
		if len(event.Targets) > 0 {
			for _, eid := range event.Targets {
				l(Event{Type: event.Type, Entity: eid, Data: event.Data})
			}
		} else {
			l(Event{Type: event.Type, Entity: event.Entity, Data: event.Data})
		}
	}
}

// 处理延迟事件
func (bus *AdvancedEventBus) Tick() {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	remain := bus.queue[:0]
	for _, e := range bus.queue {
		if e.Delay > 0 {
			e.Delay -= 10 * time.Millisecond // 假设每帧tick 10ms
			if e.Delay <= 0 {
				listeners := append([]EventListener{}, bus.listeners[e.Type]...)
				bus.lock.Unlock()
				bus.dispatchWithListeners(e, listeners)
				bus.lock.Lock()
			} else {
				remain = append(remain, e)
			}
		}
	}
	bus.queue = remain
}

func (bus *AdvancedEventBus) Chain(event CustomEvent, next CustomEvent) {
	bus.Publish(event)
	bus.Publish(next)
}

type EventFilter func(Event) bool

type EventSubscription struct {
	Priority int
	Filter   EventFilter
	Handler  EventListener
}

type PriorityEventBus struct {
	subs map[EventType][]EventSubscription
	lock sync.RWMutex
}

func NewPriorityEventBus() *PriorityEventBus {
	return &PriorityEventBus{subs: make(map[EventType][]EventSubscription)}
}

func (bus *PriorityEventBus) Subscribe(eventType EventType, sub EventSubscription) {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	bus.subs[eventType] = append(bus.subs[eventType], sub)
	// 按优先级排序
	subs := bus.subs[eventType]
	for i := len(subs) - 1; i > 0; i-- {
		if subs[i].Priority > subs[i-1].Priority {
			subs[i], subs[i-1] = subs[i-1], subs[i]
		}
	}
}

func (bus *PriorityEventBus) Publish(event Event) {
	bus.lock.RLock()
	subs := append([]EventSubscription{}, bus.subs[event.Type]...)
	bus.lock.RUnlock()
	for _, sub := range subs {
		if sub.Filter == nil || sub.Filter(event) {
			sub.Handler(event)
		}
	}
}

type EventLog []Event

func (bus *AdvancedEventBus) SaveLog() EventLog {
	bus.lock.RLock()
	defer bus.lock.RUnlock()
	log := make(EventLog, 0, len(bus.queue))
	for _, e := range bus.queue {
		log = append(log, Event{Type: e.Type, Entity: e.Entity, Data: e.Data})
	}
	return log
}

func (bus *AdvancedEventBus) Replay(log EventLog) {
	for _, e := range log {
		bus.Publish(CustomEvent{Type: e.Type, Entity: e.Entity, Data: e.Data})
	}
}
