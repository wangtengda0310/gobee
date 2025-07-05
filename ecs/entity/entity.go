// Package entity provide ecs entity management
//
// entity id stands for an entity index in the Pool and may reusable after delete
//
// entity value is a bitset of component types
package entity

import (
	"github.com/wangtengda0310/gobee/ecs/component"
)

// Pool is a fixed size array to store entity data
// Pool[0] is the next entity id to allocate
// Pool[1:L] stores the entity data, where index stands for entity id
// Pool[m:R] stores removed entity ids to reuse, where m is the last removed entity index
var Pool = make([]Archetype, 256)

var (
	L = 1             // stands for next allocate
	M = len(Pool) - 1 // collects removed entity to reuse
	R = M             // capacity for Pool
)

type Entity uint64

func init() {
	Pool[0] = 0
}

var Chunks = map[Archetype]*Chunk{}

func New(components ...component.Component) Entity {
	// archetype stands for all components that entity holding by bitset
	var archetype int
	for _, c := range components {
		archetype = archetype | int(c.Type())
	}

	Chunks[Archetype(archetype)] = &Chunk{Components: map[component.Type][]component.Component{}}
	for _, c := range components {
		cs := Chunks[Archetype(archetype)].Components[c.Type()]
		Chunks[Archetype(archetype)].Components[c.Type()] = append(cs, c)
	}

	if M == R {
		// new entity
		e := Pool[0] + 1
		Pool[L] = Archetype(archetype)
		L++
		Pool[0] = e
		return Entity(L - 1)
	} else {
		// reuse entity
		//M++
		//e := Pool[M]
		Pool[L] = Archetype(archetype)
		L++
		return Entity(L - 1)
	}
}

func Del(e Entity) {
	L--
	Pool[e], Pool[M] = Pool[L], Pool[e]
	M--
}
func Components(e Entity) (r []component.Component) {
	return nil
}
func AddComponent(e Entity, components ...component.Component) {
	var v = Pool[e]
	for _, c := range components {
		i := int(c.Type())
		v = Archetype(int(v) | i)
	}
	Pool[e] = v
	component.AddComponent(components...)
}

type Archetype int
type Chunk struct {
	Archetype  Archetype
	Components map[component.Type][]component.Component
}
