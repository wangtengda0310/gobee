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
		func() componentA { component.AddComponent(a); return a },
		func() componentB { component.AddComponent(b); return b },
		func() componentC { component.AddComponent(c); return c },
	)
	s1.update()
	assert.Equal(t, []any{a, b, c}, r)

	r = nil
	Group(func() {
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

func TestOne(t *testing.T) {
	One(1, func() {})
	assert.Equal(t, [][]int{{1}}, structDispatcher)
	One(2, func() {})
	assert.Equal(t, [][]int{{1}, {2}}, structDispatcher)
	One(4, func() {})
	assert.Equal(t, [][]int{{1}, {2}, {4}}, structDispatcher)

	c := Range(*structDispatcher)
	assert.Equal(t, 3, c, "3个不相关的system应该可以并行执行")
}

func TestGroup(t *testing.T) {
	Group(nil, 1, 2)
	assert.Equal(t, [][]int{{3}}, structDispatcher)
	Group(nil, 4, 8)
	assert.Equal(t, [][]int{{3}, {12}}, structDispatcher)
	Group(nil, 1, 16)
	assert.Equal(t, [][]int{{19}, {12}}, structDispatcher)
	Group(nil, 1, 4)
	assert.Equal(t, [][]int{{31}}, structDispatcher)

	entity.New(1, 2, 4, 8, 16, 32, 64)

	c := Range(*structDispatcher)
	assert.Equal(t, 1, c, "1个相关的system应该可以并行执行")
	Group(nil, 32, 64)
	assert.Equal(t, [][]int{{31}, {96}}, structDispatcher)
	c = Range(*structDispatcher)
	assert.Equal(t, 2, len(structDispatcher.group), "2个不相关的system应该可以并行执行")
}

func TestRound(t *testing.T) {
	t.Skipf("不重要的功能，先不测试")
	f := func() {}
	Round(1, 2, f)
	assert.Equal(t, &systemDispatcher{group: []DispatcherGroup{{
		round: []DispatcherRound{{
			archetype: 1,
			system:    []func(){f},
		}, {
			archetype: 2,
			system:    []func(){f},
		}},
	}}}, structDispatcher)
	Round(4, 8, f)
	assert.Equal(t, &systemDispatcher{group: []DispatcherGroup{{
		round: []DispatcherRound{{
			archetype: 1,
			system:    []func(){f},
		}, {
			archetype: 2,
			system:    []func(){f},
		}},
	}, {
		round: []DispatcherRound{{
			archetype: 4,
			system:    []func(){f},
		}, {
			archetype: 8,
			system:    []func(){f},
		}},
	}}}, structDispatcher)
	Round(1, 8, f)
	assert.Equal(t, &systemDispatcher{group: []DispatcherGroup{{
		round: []DispatcherRound{{
			archetype: 1,
			system:    []func(){f},
		}, {
			archetype: 2 | 8,
			system:    []func(){f, f},
		}},
	}, {
		round: []DispatcherRound{{
			archetype: 4,
			system:    []func(){f},
		}, {
			archetype: 8,
			system:    []func(){f},
		}},
	}}}, structDispatcher)
	Round(4, 16, f)
	assert.Equal(t, &systemDispatcher{group: []DispatcherGroup{{
		round: []DispatcherRound{{
			archetype: 1,
			system:    []func(){f},
		}, {
			archetype: 2 | 8,
			system:    []func(){f, f},
		}},
	}, {
		round: []DispatcherRound{{
			archetype: 4,
			system:    []func(){f},
		}, {
			archetype: 8 | 16,
			system:    []func(){f, f},
		}},
	}}}, structDispatcher)
}

func TestRange(t *testing.T) {
	tests := []struct {
		name       string
		pool       []int
		l, m, r    int
		components [][]int
		dispatcher systemDispatcher
		want       int
	}{
		{
			name:       "没有任何entity挂载组件",
			pool:       []int{0, 0, 0, 0},
			l:          4,
			m:          255,
			r:          255,
			components: [][]int{nil, {1}, {2}, {4}},
			dispatcher: systemDispatcher{group: []DispatcherGroup{{
				round: []DispatcherRound{{
					archetype: 1,
					system:    nil,
				}},
			}, {
				round: []DispatcherRound{{
					archetype: 2,
					system:    nil,
				}},
			}, {
				round: []DispatcherRound{{
					archetype: 4,
					system:    nil,
				}},
			}}},
			want: 0,
		},
		{
			name:       "3个system分别依赖component[1], component[2], component[4]",
			pool:       []int{0, 1, 2, 4},
			l:          4,
			m:          255,
			r:          255,
			components: [][]int{nil, {1}, {2}, {4}},
			dispatcher: systemDispatcher{group: []DispatcherGroup{{
				round: []DispatcherRound{{
					archetype: 1,
					system:    nil,
				}},
			}, {
				round: []DispatcherRound{{
					archetype: 2,
					system:    nil,
				}},
			}, {
				round: []DispatcherRound{{
					archetype: 4,
					system:    nil,
				}},
			}}},
			want: 3,
		},
		{
			name:       "3个system分别依赖component[1+2=3], component[1+4=5], component[2+4=6]",
			pool:       []int{0, 3, 5, 6},
			l:          4,
			m:          255,
			r:          255,
			components: [][]int{nil, {3}, {5}, {6}},
			dispatcher: systemDispatcher{group: []DispatcherGroup{{
				round: []DispatcherRound{{
					archetype: 3,
					system:    nil,
				}},
			}, {
				round: []DispatcherRound{{
					archetype: 5,
					system:    nil,
				}},
			}, {
				round: []DispatcherRound{{
					archetype: 6,
					system:    nil,
				}},
			}}},
			want: 3,
		},
		{
			name:       "一个entity挂载的component满足3个system的依赖",
			pool:       []int{0, 7},
			l:          2,
			m:          255,
			r:          255,
			components: [][]int{nil, {1}, {2}, {4}},
			dispatcher: systemDispatcher{group: []DispatcherGroup{{
				round: []DispatcherRound{{
					archetype: 1,
					system:    nil,
				}},
			}, {
				round: []DispatcherRound{{
					archetype: 2,
					system:    nil,
				}},
			}, {
				round: []DispatcherRound{{
					archetype: 4,
					system:    nil,
				}},
			}}},
			want: 3,
		},
		{
			name:       "entity挂载的component不满足system的依赖",
			pool:       []int{0, 3},
			l:          2,
			m:          255,
			r:          255,
			components: [][]int{nil, {1}, {2}},
			dispatcher: systemDispatcher{group: []DispatcherGroup{{
				round: []DispatcherRound{{
					archetype: 6,
					system:    nil,
				}},
			}}},
			want: 0,
		},
		{
			name:       "system不依赖component",
			pool:       []int{0, 3},
			l:          2,
			m:          255,
			r:          255,
			components: [][]int{nil, {1}},
			dispatcher: systemDispatcher{group: []DispatcherGroup{{
				round: []DispatcherRound{{
					archetype: 0,
					system:    nil,
				}},
			}}},
			want: 1,
		},
		{
			name:       "一个system[1]update完以后再另一个system[3]update",
			pool:       []int{0, 3},
			l:          2,
			m:          255,
			r:          255,
			components: [][]int{nil, {1}},
			dispatcher: systemDispatcher{group: []DispatcherGroup{{
				round: []DispatcherRound{{
					archetype: 1,
					system:    nil,
				}},
			}, {
				round: []DispatcherRound{{
					archetype: 3,
					system:    nil,
				}},
			}}},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity.Pool = tt.pool
			entity.L, entity.M, entity.R = tt.l, tt.m, tt.r
			assert.Equalf(t, tt.want, Range(tt.dispatcher), "Range(%v)", tt.dispatcher)
			entity.Pool = make([]int, 256)

			// roll back
			entity.L = 1
			entity.M = len(entity.Pool) - 1
			entity.R = entity.M
		})
	}
}
