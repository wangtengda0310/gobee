package entity

import (
	"github.com/wangtengda0310/gobee/ecs/component"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert.Equal(t, 1, L)
	assert.Equal(t, 255, R)
	assert.Equal(t, 255, M)
	assert.Equal(t, 0, Pool[0])

	e := New(1)
	assert.Equal(t, 2, L)
	assert.Equal(t, 255, R)
	assert.Equal(t, 255, M)
	assert.Equal(t, Entity(1), e)
	assert.Equal(t, 1, Pool[0])
	assert.Equal(t, 1, Pool[1])

	e = New(1, 2)
	assert.Equal(t, 3, L)
	assert.Equal(t, 255, R)
	assert.Equal(t, 255, M)
	assert.Equal(t, Entity(2), e)
	assert.Equal(t, 2, Pool[0])
	assert.Equal(t, 1, Pool[1])
	assert.Equal(t, 1|2, Pool[2])

	e = New(1, 2, 4)
	assert.Equal(t, 4, L)
	assert.Equal(t, 255, R)
	assert.Equal(t, 255, M)
	assert.Equal(t, Entity(3), e)
	assert.Equal(t, 3, Pool[0])
	assert.Equal(t, 1, Pool[1])
	assert.Equal(t, 1|2, Pool[2])
	assert.Equal(t, 1|2|4, Pool[3])

	Del(2)
	assert.Equal(t, 3, L)
	assert.Equal(t, 255, R)
	assert.Equal(t, 254, M)
	assert.Equal(t, 3, Pool[0])
	assert.Equal(t, 1, Pool[1])
	assert.Equal(t, 1|2|4, Pool[2])
	assert.Equal(t, 2, Pool[255])

	e = New(1, 2, 4)
	assert.Equal(t, 4, L)
	assert.Equal(t, 255, R)
	assert.Equal(t, 255, M)
	assert.Equal(t, Entity(2), e)
	assert.Equal(t, 3, Pool[0])
	assert.Equal(t, 1, Pool[1])
	assert.Equal(t, 1|2|4, Pool[2])
	assert.Equal(t, 1|2|4, Pool[3])

	Del(2)
	Del(1)
	Del(3)

	e = New(4)
	assert.Equal(t, Entity(3), e)
	assert.Equal(t, 3, Pool[0])
	assert.Equal(t, 4, Pool[1])
	e = New(8)
	assert.Equal(t, Entity(1), e)
	assert.Equal(t, 3, Pool[0])
	assert.Equal(t, 4, Pool[1])
	assert.Equal(t, 8, Pool[2])
	e = New(64)
	assert.Equal(t, Entity(2), e)
	assert.Equal(t, 3, Pool[0])
	assert.Equal(t, 4, Pool[1])
	assert.Equal(t, 8, Pool[2])
	assert.Equal(t, 64, Pool[3])
	e = New(128)
	assert.Equal(t, Entity(4), e)
	assert.Equal(t, 4, Pool[0])
	assert.Equal(t, 4, Pool[1])
	assert.Equal(t, 8, Pool[2])
	assert.Equal(t, 64, Pool[3])
	assert.Equal(t, 128, Pool[4])
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
	assert.Equal(t, 2, Pool[e])
	AddComponent(e, &c{t: 1})
	assert.Equal(t, 3, Pool[e])
	AddComponent(e, &c{t: 4})
	assert.Equal(t, 7, Pool[e])
}
