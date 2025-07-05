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
	AddComponent(e, &c{t: 2})
	assert.EqualValues(t, 2, Pool[e])
	AddComponent(e, &c{t: 1})
	assert.EqualValues(t, 3, Pool[e])
	AddComponent(e, &c{t: 4})
	assert.EqualValues(t, 7, Pool[e])
}
