package datatypes

import (
	"math/rand"
	"testing"
	"time"
)

func TestSortedSet(t *testing.T) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	sl := NewSortedSet()

	// Test adding elements
	sl.Add("Alice", 100)
	sl.Add("Bob", 75)
	sl.Add("Charlie", 85)
	sl.Add("Diana", 95)

	// Test quick lookup by name
	if score, found := sl.GetScore("Charlie"); !found || score != 85 {
		t.Errorf("Expected score 85 for Charlie, got %d, found: %v", score, found)
	}

	if score, found := sl.GetScore("Eve"); found {
		t.Errorf("Expected no entry for Eve, but found score %d", score)
	}

	// Test removing elements
	sl.Remove("Charlie")
	if _, found := sl.GetScore("Charlie"); found {
		t.Errorf("Expected Charlie to be removed, but still found")
	}

	sl.Remove("Alice")
	if _, found := sl.GetScore("Alice"); found {
		t.Errorf("Expected Alice to be removed, but still found")
	}

	// Test adding more elements after removal
	sl.Add("Eve", 105)
	if score, found := sl.GetScore("Eve"); !found || score != 105 {
		t.Errorf("Expected score 105 for Eve, got %d, found: %v", score, found)
	}
}
