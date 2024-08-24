package datatypes

import "testing"

func TestSet_Add_Remove(t *testing.T) {
	set := NewSet[string]()

	set.Add("mango")
	if set.Size() != 1 {
		t.Errorf("Set.Size() = %d, want 1", set.Size())
	}

	if !set.Contains("mango") {
		t.Errorf("Set.Contains() = false, want true")
	}

	set.Add("banana")
	if set.Size() != 2 {
		t.Errorf("Set.Size() = %d, want 2", set.Size())
	}

	if !set.Contains("banana") {
		t.Errorf("Set.Contains() = false, want true")
	}

	set.Add("banana")
	if set.Size() != 2 {
		t.Errorf("Set.Size() = %d, want 2", set.Size())
	}

	set.Remove("banana")
	if set.Size() != 1 {
		t.Errorf("Set.Size() = %d, want 1", set.Size())
	}

	set.Remove("mango")
	if set.Size() != 0 {
		t.Errorf("Set.Size() = %d, want 0", set.Size())
	}

	set.AddMany([]string{"mango", "banana"})
	if set.Size() != 2 {
		t.Errorf("Set.Size() = %d, want 2", set.Size())
	}
}

func TestSet_Intersection(t *testing.T) {
	s1 := NewSet[string]()
	s2 := NewSet[string]()

	s1.Add("mango")
	s1.Add("banana")
	s1.Add("papaya")
	s1.Add("orange")

	s2.Add("banana")
	s2.Add("papaya")

	intersection := s1.Intersection(s2)

	if intersection.Size() != 2 {
		t.Errorf("intersection.Size() = %d, want 2", intersection.Size())
	}

	i := NewSet[string]()
	i.Add("banana")
	i.Add("papaya")

	if !intersection.Equals(i) {
		t.Errorf("intersection = %v, want %v", i, intersection)
	}
}

func TestSet_Union(t *testing.T) {
	s1 := NewSet[string]()
	s2 := NewSet[string]()

	s1.Add("mango")
	s1.Add("banana")
	s2.Add("papaya")
	s2.Add("orange")

	union := s1.Union(s2)

	if union.Size() != 4 {
		t.Errorf("intersection.Size() = %d, want 2", union.Size())
	}

	u := NewSet[string]()
	u.Add("banana")
	u.Add("papaya")
	u.Add("orange")
	u.Add("mango")

	if !union.Equals(u) {
		t.Errorf("intersection = %v, want %v", u, union)
	}
}
