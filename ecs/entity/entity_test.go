package entity

import (
	"github.com/wangtengda0310/gobee/ecs/component"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert.Equal(t, 1, l)
	assert.Equal(t, 255, r)
	assert.Equal(t, 255, m)
	assert.Equal(t, 0, pool[0])

	e := New(1)
	assert.Equal(t, 2, l)
	assert.Equal(t, 255, r)
	assert.Equal(t, 255, m)
	assert.Equal(t, Entity(1), e)
	assert.Equal(t, 1, pool[0])
	assert.Equal(t, 1, pool[1])

	e = New(1, 2)
	assert.Equal(t, 3, l)
	assert.Equal(t, 255, r)
	assert.Equal(t, 255, m)
	assert.Equal(t, Entity(2), e)
	assert.Equal(t, 2, pool[0])
	assert.Equal(t, 1, pool[1])
	assert.Equal(t, 1|2, pool[2])

	e = New(1, 2, 4)
	assert.Equal(t, 4, l)
	assert.Equal(t, 255, r)
	assert.Equal(t, 255, m)
	assert.Equal(t, Entity(3), e)
	assert.Equal(t, 3, pool[0])
	assert.Equal(t, 1, pool[1])
	assert.Equal(t, 1|2, pool[2])
	assert.Equal(t, 1|2|4, pool[3])

	Del(2)
	assert.Equal(t, 3, l)
	assert.Equal(t, 255, r)
	assert.Equal(t, 254, m)
	assert.Equal(t, 3, pool[0])
	assert.Equal(t, 1, pool[1])
	assert.Equal(t, 1|2|4, pool[2])
	assert.Equal(t, 2, pool[255])

	e = New(1, 2, 4)
	assert.Equal(t, 4, l)
	assert.Equal(t, 255, r)
	assert.Equal(t, 255, m)
	assert.Equal(t, Entity(2), e)
	assert.Equal(t, 3, pool[0])
	assert.Equal(t, 1, pool[1])
	assert.Equal(t, 1|2|4, pool[2])
	assert.Equal(t, 1|2|4, pool[3])

	Del(2)
	Del(1)
	Del(3)

	e = New(4)
	assert.Equal(t, Entity(3), e)
	assert.Equal(t, 3, pool[0])
	assert.Equal(t, 4, pool[1])
	e = New(8)
	assert.Equal(t, Entity(1), e)
	assert.Equal(t, 3, pool[0])
	assert.Equal(t, 4, pool[1])
	assert.Equal(t, 8, pool[2])
	e = New(64)
	assert.Equal(t, Entity(2), e)
	assert.Equal(t, 3, pool[0])
	assert.Equal(t, 4, pool[1])
	assert.Equal(t, 8, pool[2])
	assert.Equal(t, 64, pool[3])
	e = New(128)
	assert.Equal(t, Entity(4), e)
	assert.Equal(t, 4, pool[0])
	assert.Equal(t, 4, pool[1])
	assert.Equal(t, 8, pool[2])
	assert.Equal(t, 64, pool[3])
	assert.Equal(t, 128, pool[4])
}

type c struct {
	t component.Type
}

func (c *c) Type() component.Type {
	return c.t
}
func TestAddComponent(t *testing.T) {
	e := New()
	AddComponent(e, &c{t: 2})
	assert.Equal(t, 2, l)
	AddComponent(e, &c{t: 1})
	assert.Equal(t, 3, l)
	AddComponent(e, &c{t: 4})
	assert.Equal(t, 7, l)
}

func TestRange(t *testing.T) {
	tests := []struct {
		name       string
		pool       []int
		l, m, r    int
		components [][]int
		dispatcher [][]int
		want       int
	}{
		{
			name:       "没有任何entity挂载组件",
			pool:       []int{0, 0, 0, 0},
			l:          4,
			m:          255,
			r:          255,
			components: [][]int{nil, {1}, {2}, {4}},
			dispatcher: [][]int{{1}, {2}, {4}},
			want:       0,
		},
		{
			name:       "3个system分别依赖component[1], component[2], component[4]",
			pool:       []int{0, 1, 2, 4},
			l:          4,
			m:          255,
			r:          255,
			components: [][]int{nil, {1}, {2}, {4}},
			dispatcher: [][]int{{1}, {2}, {4}},
			want:       3,
		},
		{
			name:       "3个system分别依赖component[1+2=3], component[1+4=5], component[2+4=6]",
			pool:       []int{0, 3, 5, 6},
			l:          4,
			m:          255,
			r:          255,
			components: [][]int{nil, {3}, {5}, {6}},
			dispatcher: [][]int{{3}, {5}, {6}},
			want:       3,
		},
		{
			name:       "一个entity挂载的component满足3个system的依赖",
			pool:       []int{0, 7},
			l:          2,
			m:          255,
			r:          255,
			components: [][]int{nil, {1}, {2}, {4}},
			dispatcher: [][]int{{1}, {2}, {4}},
			want:       3,
		},
		{
			name:       "entity挂载的component不满足system的依赖",
			pool:       []int{0, 3},
			l:          2,
			m:          255,
			r:          255,
			components: [][]int{nil, {1}, {2}},
			dispatcher: [][]int{{6}},
			want:       0,
		},
		{
			name:       "system不依赖component",
			pool:       []int{0, 3},
			l:          2,
			m:          255,
			r:          255,
			components: [][]int{nil, {1}},
			dispatcher: [][]int{{0}},
			want:       1,
		},
		{
			name:       "一个system[1]update完以后再另一个system[3]update",
			pool:       []int{0, 3},
			l:          2,
			m:          255,
			r:          255,
			components: [][]int{nil, {1}},
			dispatcher: [][]int{{1, 3}},
			want:       2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool = tt.pool
			l, m, r = tt.l, tt.m, tt.r
			assert.Equalf(t, tt.want, Range(tt.dispatcher), "Range(%v)", tt.dispatcher)
		})
	}
}
