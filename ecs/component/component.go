package component

var dense = []int{}
var sparse = []int{}

// Type represents a component type in the ECS system. hold only one flag in bits
type Type int
type Component interface {
	Type() Type
}
