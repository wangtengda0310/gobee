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

	e := entity.New(typeA, typeB, typeC)
	t.Log(e)

	var r []any
	s1cb := func(v1 componentA, v2 componentB, v3 componentC) {
		r = append(r, v1)
		r = append(r, v2)
		r = append(r, v3)
	}
	s1 := New[componentA, componentB, componentC](
		s1cb,
		func() componentA { return a },
		func() componentB { return b },
		func() componentC { return c },
	)
	s1.update()
	assert.Equal(t, []any{a, b, c}, r)

	r = nil
	s2 := func(v1 componentA, v2 componentC) {
		r = append(r, v1)
		r = append(r, v2)
	}
	Update(s2)
	assert.Equal(t, []any{a, b}, r)

	// s3 := New(b, c)
	// s4 := New(a)
	// s5 := New(b)
	// s6 := New(c)
}

type S interface {
	update()
}
type s struct {
}

func (s s) update() {}

type i[TC component.Type] interface {
	Type() TC
}

func New[A, B, C component.Component](cb func(v1 A, v2 B, v3 C), t1 func() A, t2 func() B, t3 func() C) S {
	cb(t1(), t2(), t3())
	return s{}
}

func Update(cb func(a componentA, c componentC)) {
	//cb(nil, nil)
}

func TestOne(t *testing.T) {
	One(1)
	assert.Equal(t, [][]int{{1}}, dispatcher)
	One(2)
	assert.Equal(t, [][]int{{1}, {2}}, dispatcher)
	One(4)
	assert.Equal(t, [][]int{{1}, {2}, {4}}, dispatcher)

	c := entity.Range(dispatcher)
	assert.Equal(t, 3, c, "3个不相关的system应该可以并行执行")
}

func TestGroup(t *testing.T) {
	Group(1, 2)
	assert.Equal(t, [][]int{{3}}, dispatcher)
	Group(4, 8)
	assert.Equal(t, [][]int{{3}, {12}}, dispatcher)
	Group(1, 16)
	assert.Equal(t, [][]int{{19}, {12}}, dispatcher)
	Group(1, 4)
	assert.Equal(t, [][]int{{31}}, dispatcher)

	entity.New(1, 2, 4, 8, 16, 32, 64)

	c := entity.Range(dispatcher)
	assert.Equal(t, 1, c, "1个相关的system应该可以并行执行")
	Group(32, 64)
	assert.Equal(t, [][]int{{31}, {96}}, dispatcher)
	c = entity.Range(dispatcher)
	assert.Equal(t, 2, len(dispatcher), "2个不相关的system应该可以并行执行")
}

func TestRound(t *testing.T) {
	Round(1, 2)
	assert.Equal(t, [][]int{{1, 2}}, dispatcher)
	Round(4, 8)
	assert.Equal(t, [][]int{{1, 2}, {4, 8}}, dispatcher)
	//Round(1, 8)
	//assert.Equal(t, [][]int{{1, {2, 8}}, {4, 8}}, dispatcher)
	//Round(4, 16)
	//assert.Equal(t, [][]int{{1, 2}, {4, {8, 16}}}, dispatcher)
}
