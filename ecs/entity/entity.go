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

func Range(dispatcher [][]int) int {
	var c int
	for i := 1; i < l; i++ {
		for _, components := range dispatcher {
			for _, i2 := range components {
				// load components
				if pool[i]&i2 == i2 {
					c++
					// not grouped can async
					go func(c int) {
						var v = 1
						for v <= i2 {
							if v&i2 == 0 {
								continue
							}
							t := component.Type(v)

							println("Entity:", i, "Components:", pool[i], "Dispatcher:", i2, "Group", c, "type:", t)
							v = v << 1
						}
					}(c)
				}
			}
		}
	}
	return c
}
