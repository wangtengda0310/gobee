package ecs

import (
	"testing"
)

const (
	Position = iota + 1
	Velocity
	Shape
)

type PositionComponent struct {
	X int
	Y int
}
type VelocityComponent struct {
	X int
	Y int
}
type ShapeComponent struct {
	Shape string
}

func TestAddComponent(t *testing.T) {
	//assert.Equal(t, 1, entity.New())
	//
	//e := entity.New(
	//	&PositionComponent{X: 100, Y: 100},
	//	&VelocityComponent{X: 1, Y: 1})
	//
	//assert.Equal(t, Position|Velocity, entity.Components(e))
	//
	//sys := system.New(component.Position, component.Velocity)
	//sys.Update(func(p *PositionComponent, v *VelocityComponent) {
	//	p.X += v.X
	//	p.Y += v.Y
	//})
	//pos := component.Get(entity, component.Position)
	//assert.Equal(t, 101, pos.X)
	//assert.Equal(t, 101, pos.Y)
	//vel := component.Get(entity, component.Velocity)
	//assert.Equal(t, 1, vel.X)
	//assert.Equal(t, 1, vel.Y)
	//
	//entity.AddComponent(component.New("shape", "circle"))
	//assert.Equal(t, 3, len(entity.Components))
}
