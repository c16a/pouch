package datatypes

import (
	"encoding/json"
	"reflect"
)

type Comparable interface {
	comparable
}

type Set[T Comparable] struct {
	Values map[T]bool `json:"values"`
	Name   string     `json:"name"`
}

func NewSet[T Comparable]() *Set[T] {
	return &Set[T]{Values: make(map[T]bool), Name: "set"}
}

func (s *Set[T]) GetName() string {
	return s.Name
}

func (s *Set[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s)
}

// AddMany adds the values to the set and returns the number of items that have actually been added.
func (s *Set[T]) AddMany(values []T) int {
	var count int
	for _, v := range values {
		count += s.Add(v)
	}
	return count
}

func (s *Set[T]) Add(value T) int {
	if !s.Contains(value) {
		s.Values[value] = true
		return 1
	}
	return 0
}

func (s *Set[T]) Remove(value T) {
	delete(s.Values, value)
}

func (s *Set[T]) Contains(value T) bool {
	_, ok := s.Values[value]
	return ok
}

func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	intersection := NewSet[T]()
	for value := range s.Values {
		if other.Contains(value) {
			intersection.Add(value)
		}
	}
	return intersection
}

func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	union := NewSet[T]()
	for value := range s.Values {
		union.Add(value)
	}
	for value := range other.Values {
		union.Add(value)
	}
	return union
}

func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	difference := NewSet[T]()
	for value := range s.Values {
		if !other.Contains(value) {
			difference.Add(value)
		}
	}
	return difference
}

func (s *Set[T]) Equals(other *Set[T]) bool {
	return reflect.DeepEqual(s, other)
}

func (s *Set[T]) Size() int {
	return len(s.Values)
}

func (s *Set[T]) GetMembers() []T {
	members := make([]T, 0, s.Size())
	for value := range s.Values {
		members = append(members, value)
	}
	return members
}

func (s *Set[T]) Copy() *Set[T] {
	set := NewSet[T]()
	set.AddMany(s.GetMembers())
	return set
}
