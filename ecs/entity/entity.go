// Package entity provide ecs entity management
//
// entity id stands for an entity index in the pool and may reusable after delete
//
// entity value is a bitset of component types
package entity

import (
	"github.com/wangtengda0310/gobee/ecs/component"
)

var pool = make([]int, 256)

var (
	l = 1             // stands for next allocate
	m = len(pool) - 1 // collects removed entity to reuse
	r = m             // capacity for pool
)

type Entity uint64

func init() {
	pool[0] = 0
}

func New(components ...component.Type) Entity {
	// v stands for all components that entity holding by bitset
	var v int
	for _, c := range components {
		i := int(c)
		v = v | i
	}

	if m == r {
		// new entity
		e := pool[0] + 1
		pool[l] = v
		l++
		pool[0] = e
		return Entity(l - 1)
	} else {
		// reuse entity
		m++
		e := pool[m]
		pool[l] = v
		l++
		return Entity(e)
	}
}

func Del(e Entity) {
	l--
	pool[e], pool[m] = pool[l], int(e)
	m--
}
func Components(e Entity) (r []component.Component) {
	return nil
}
func AddComponent(e Entity, components ...component.Component) {
	var v = pool[e]
	for _, c := range components {
		i := int(c.Type())
		v = v | i
	}
	pool[e] = v
}
