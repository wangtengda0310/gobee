package entity

import (
	"github.com/wangtengda0310/gobee/ecs/component"
	"testing"

	"github.com/stretchr/testify/assert"
)

type c struct {
	t component.Type
}

func (c *c) Type() component.Type {
	return c.t
}

func TestAddComponent(t *testing.T) {
	e := New()
	assert.EqualValues(t, 0, Chunks[0].Archetype)

	AddComponent(e, &c{t: 1})
	assert.EqualValues(t, 0, Chunks[0].Len())
	assert.EqualValues(t, 1, Chunks[1].Archetype)
	assert.EqualValues(t, 1, Pool[e])

	AddComponent(e, &c{t: 2})
	assert.EqualValues(t, 0, Chunks[0].Len())
	assert.EqualValues(t, 0, Chunks[1].Len())
	assert.EqualValues(t, 3, Chunks[3].Archetype)
	assert.EqualValues(t, 3, Pool[e])

	AddComponent(e, &c{t: 4})
	assert.EqualValues(t, 0, Chunks[0].Len())
	assert.EqualValues(t, 0, Chunks[1].Len())
	assert.EqualValues(t, 0, Chunks[3].Len())
	assert.EqualValues(t, 7, Chunks[7].Archetype)
	assert.EqualValues(t, 7, Pool[e])
}

func TestRemoveComponent(t *testing.T) {
	e := New(&c{t: 1}, &c{t: 2}, &c{t: 4})
	assert.EqualValues(t, 7, Chunks[7].Archetype)
	assert.EqualValues(t, 3, Chunks[7].Len())

	RemoveComponent(e, 1)
	assert.EqualValues(t, 0, Chunks[7].Len())
	assert.EqualValues(t, 6, Chunks[6].Archetype)
	assert.EqualValues(t, 2, Chunks[6].Len())
	assert.EqualValues(t, 6, Pool[e])

	RemoveComponent(e, 2)
	assert.EqualValues(t, 0, Chunks[7].Len())
	assert.EqualValues(t, 0, Chunks[6].Len())
	assert.EqualValues(t, 1, Chunks[4].Len())
	assert.EqualValues(t, 4, Chunks[4].Archetype)
	assert.EqualValues(t, 4, Pool[e])

	RemoveComponent(e, 4)
	assert.EqualValues(t, 0, Chunks[7].Len())
	assert.EqualValues(t, 0, Chunks[6].Len())
	assert.EqualValues(t, 0, Chunks[4].Len())
	assert.EqualValues(t, 0, Chunks[0].Len())
	assert.EqualValues(t, 0, Chunks[0].Archetype)
	assert.EqualValues(t, 0, Pool[e])
}
