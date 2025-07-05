package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wangtengda0310/gobee/ecs/component"
	"github.com/wangtengda0310/gobee/ecs/entity"
)

const (
	typeA component.Type = iota
	typeB component.Type = iota
	typeC component.Type = iota
)

type componentA struct {
	value int
}

func (c componentA) Type() component.Type {
	return typeA
}

type componentB struct {
	value int
}

func (c componentB) Type() component.Type {
	return typeB
}

type componentC struct {
	value int
}

func (c componentC) Type() component.Type {
	return typeC
}
func TestNew(t *testing.T) {
	a := componentA{1}
	b := componentB{2}
	c := componentC{3}

	e := entity.New(a, b, c)
	t.Log(e)

	var r []any
	s1cb := func(v1 componentA, v2 componentB, v3 componentC) {
		r = append(r, v1)
		r = append(r, v2)
		r = append(r, v3)
	}
	s1 := New[componentA, componentB, componentC](
		s1cb,
		func() componentA { component.AddComponent(a); return a },
		func() componentB { component.AddComponent(b); return b },
		func() componentC { component.AddComponent(c); return c },
	)
	s1.update()
	assert.Equal(t, []any{a, b, c}, r)

	r = nil
	Group(func(chunk entity.Chunk) {
		r = append(r, a)
		r = append(r, b)
	}, typeA, typeB)
	Range(*structDispatcher)
	//Update(e, s2)
	assert.Equal(t, []any{a, b}, r)

	// s3 := New(b, c)
	// s4 := New(a)
	// s5 := New(b)
	// s6 := New(c)
	structDispatcher = &systemDispatcher{}
}

type S interface {
	update()
}
type s struct {
	systems []func()
}

func (s s) update() {
	for _, f := range s.systems {
		f() // 假设1, 2, 3是A, B, C的实例
	}
}

type i[TC component.Type] interface {
	Type() TC
}

func New[A, B, C component.Component](cb func(v1 A, v2 B, v3 C), t1 func() A, t2 func() B, t3 func() C) S {
	return s{[]func(){func() {
		cb(t1(), t2(), t3())
	}}}
}

func Update(e entity.Entity, cb func(componentA, componentC)) {
}

func TestGroup(t *testing.T) {
	Group(func(chunk entity.Chunk) {}, 1, 2)
	assert.Equal(t, 3, structDispatcher.group[0].round[0].archetype)
	Group(func(chunk entity.Chunk) {}, 4, 8)
	assert.Equal(t, 12, structDispatcher.group[1].round[0].archetype)
	Group(func(chunk entity.Chunk) {}, 1, 16)
	assert.Equal(t, 17, structDispatcher.group[2].round[0].archetype)
	Group(func(chunk entity.Chunk) {}, 1, 4)
	assert.Equal(t, 5, structDispatcher.group[3].round[0].archetype)

	entity.New(componentA{1}, componentB{1}, componentC{1})

	c := Range(*structDispatcher)
	assert.Equal(t, 1, c, "1个相关的system应该可以并行执行")
	Group(nil, 32, 64)
	assert.Equal(t, 96, structDispatcher.group[4].round[0].archetype)
	c = Range(*structDispatcher)
	assert.Equal(t, 5, len(structDispatcher.group), "2个不相关的system应该可以并行执行")
}
