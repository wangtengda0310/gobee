package component

type SparseSet[T Component] struct {
	dense  [][]T
	sparse []int
}

func (s *SparseSet[T]) Get(id int) []T {
	if id < 0 || id >= len(s.sparse) {
		return nil
	}
	idx := s.sparse[id]
	if idx < 0 || idx >= len(s.dense) {
		return nil
	}
	return s.dense[idx]
}

func (s *SparseSet[T]) Add(id int, c T) {
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
func (s *SparseSet[T]) Del(id int) {
	if id < 0 || id >= len(s.sparse) {
		return
	}
	idx := s.sparse[id]
	if idx < 0 || idx >= len(s.dense) {
		return
	}

	s.dense[len(s.dense)-1] = s.dense[idx]
	s.dense = s.dense[:len(s.dense)-1]

	s.sparse[id] = -1 // mark as empty

}

// Type represents a component type in the ECS system. hold only one flag in bits
type Type int
type Component interface {
	Type() Type
}
