// Package entity provide ecs entity management
//
// entity id stands for an entity index in the Pool and may reusable after delete
//
// entity value is a bitset of component types
package entity

import (
	"github.com/wangtengda0310/gobee/ecs/component"
)

var id Entity
var Pool = map[Entity]Archetype{}

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

	id++
	Pool[id] = Archetype(archetype)

	Chunks[Archetype(archetype)] = &Chunk{Components: map[component.Type]component.SparseSet[component.Component]{}}
	for _, c := range components {
		sparse := Chunks[Archetype(archetype)].Components[c.Type()]
		sparse.Add(int(id), c)
	}

	return id
}

func Del(e Entity) {
	delete(Pool, e)
	if chunk, ok := Chunks[Pool[e]]; ok {
		for _, cs := range chunk.Components {
			cs.Del(int(e))
		}
		delete(Chunks, Pool[e])
	}
}
func RemoveComponent(e Entity, components ...component.Component) (r []component.Component) {
	var v = Pool[e]
	for _, c := range Chunks[v].Components {
		c.Del(int(e))
	}
	for _, c := range components {
		i := int(c.Type())
		v = Archetype(int(v) ^ i)
	}
	Pool[e] = v

	for _, c := range components {
		c2, ok := Chunks[v].Components[c.Type()]
		if !ok {
			continue
		}
		c2.Del(int(id))
	}

	return nil
}
func AddComponent(e Entity, components ...component.Component) {
	var v = Pool[e]
	for _, c := range Chunks[v].Components {
		c.Del(int(e))
	}
	for _, c := range components {
		i := int(c.Type())
		v = Archetype(int(v) | i)
	}
	Pool[e] = v
	if Chunks[v] == nil {
		Chunks[v] = &Chunk{Archetype: v, Components: map[component.Type]component.SparseSet[component.Component]{}}
	}
	for _, c := range components {
		c2, ok := Chunks[v].Components[c.Type()]
		if !ok {
			c2 = component.SparseSet[component.Component]{}
			Chunks[v].Components[c.Type()] = c2
		}
		c2.Add(int(id), c)
	}
}

type Archetype int
type Chunk struct {
	Archetype  Archetype
	Components map[component.Type]component.SparseSet[component.Component]
}
