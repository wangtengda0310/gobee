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
