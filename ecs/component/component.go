package component

type sparseSet[T Component] struct {
	dense  [][]T
	sparse []int
}

var allType = sparseSet[Component]{}

func (s *sparseSet[T]) Get(id int) []T {
	if id < 0 || id >= len(s.sparse) {
		return nil
	}
	idx := s.sparse[id]
	if idx < 0 || idx >= len(s.dense) {
		return nil
	}
	return s.dense[idx]
}

func (s *sparseSet[T]) Add(id int, c T) {
	if s.sparse == nil {
		s.sparse = make([]int, len(s.dense))
	}
	if id > len(s.sparse)-1 {
		// expand sparse array
		newSparse := make([]int, id+1)
		copy(newSparse, s.sparse)
		s.sparse = newSparse
	}
	var idx = s.sparse[id]
	if idx > 0 {
		s.dense[idx] = append(s.dense[idx], c)
		return
	}
	idx = len(s.dense)
	s.dense = append(s.dense, nil)
	s.dense[idx] = append(s.dense[idx], c)
	s.sparse[id] = idx
}

// Type represents a component type in the ECS system. hold only one flag in bits
type Type int
type Component interface {
	Type() Type
}

func AddComponent(components ...Component) {
	for _, c := range components {
		t := int(c.Type())
		allType.Add(t, c)
	}
}

func Data(c Type) []Component {
	return allType.Get(int(c))
}
