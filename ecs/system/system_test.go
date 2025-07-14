package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wangtengda0310/gobee/ecs/component"
	"github.com/wangtengda0310/gobee/ecs/entity"
)

const (
	typeA component.Type = 1
	typeB component.Type = 2
	typeC component.Type = 4
	typeD component.Type = 8
	typeE component.Type = 16
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

	e := entity.New(a)
	t.Log(e)
	assert.EqualValues(t, 1, entity.Chunks[1].Archetype)

	e = entity.New(b)
	t.Log(e)
	assert.EqualValues(t, 2, entity.Chunks[2].Archetype)

	e = entity.New(c)
	t.Log(e)
	assert.EqualValues(t, 4, entity.Chunks[4].Archetype)

	e = entity.New(a, b)
	t.Log(e)
	assert.EqualValues(t, 3, entity.Chunks[3].Archetype)

	e = entity.New(b, c)
	t.Log(e)
	assert.EqualValues(t, 6, entity.Chunks[6].Archetype)

	e = entity.New(a, c)
	t.Log(e)
	assert.EqualValues(t, 5, entity.Chunks[5].Archetype)

	e = entity.New(a, b, c)
	t.Log(e)
	assert.EqualValues(t, 7, entity.Chunks[7].Archetype)

	var called int
	Group(func(chunk entity.Chunk) {
		called++
	}, typeA)
	Update()
	assert.EqualValues(t, 4, called)

	structDispatcher = &systemDispatcher{}
}

func TestGroup(t *testing.T) {
	Group(func(chunk entity.Chunk) {}, typeA, typeB)
	assert.Equal(t, 3, structDispatcher.group[0].round[0].archetype)
	Group(func(chunk entity.Chunk) {}, typeC, typeD)
	assert.Equal(t, 12, structDispatcher.group[1].round[0].archetype)
	Group(func(chunk entity.Chunk) {}, typeA, typeE)
	assert.Equal(t, 17, structDispatcher.group[2].round[0].archetype)
	Group(func(chunk entity.Chunk) {}, typeA, typeC)
	assert.Equal(t, 5, structDispatcher.group[3].round[0].archetype)

	entity.New(componentA{1}, componentB{1}, componentC{1})

	c := Update()
	assert.Equal(t, 1, c, "1个相关的system应该可以并行执行")
	Group(nil, 32, 64)
	assert.Equal(t, 96, structDispatcher.group[4].round[0].archetype)
	c = Update()
	assert.Equal(t, 5, len(structDispatcher.group), "2个不相关的system应该可以并行执行")
}
