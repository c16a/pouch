package datatypes

import (
	"math/rand"
)

// MaxLevel is the maximum level for the skip list nodes.
const MaxLevel = 16

// Probability is the probability used to decide whether to increase the level of a node.
const Probability = 0.5

// sortedSetNode represents a node in the skip list.
type sortedSetNode struct {
	name  string
	score int
	next  []*sortedSetNode
}

// SortedSet represents the skip list structure.
type SortedSet struct {
	header *sortedSetNode
	level  int
	table  map[string]*sortedSetNode // Hash table for quick lookup by name
}

// newSortedSetNode creates a new node with a given name, score, and level.
func newSortedSetNode(name string, score, level int) *sortedSetNode {
	return &sortedSetNode{
		name:  name,
		score: score,
		next:  make([]*sortedSetNode, level),
	}
}

// NewSortedSet creates a new skip list.
func NewSortedSet() *SortedSet {
	return &SortedSet{
		header: newSortedSetNode("", 0, MaxLevel),
		level:  1,
		table:  make(map[string]*sortedSetNode),
	}
}

// randomLevel generates a random level for a new node.
func randomLevel() int {
	level := 1
	for rand.Float64() < Probability && level < MaxLevel {
		level++
	}
	return level
}

// Add inserts a new element (name, score) into the skip list.
func (sl *SortedSet) Add(name string, score int) {
	update := make([]*sortedSetNode, MaxLevel)
	current := sl.header

	for i := sl.level - 1; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].score < score {
			current = current.next[i]
		}
		update[i] = current
	}

	level := randomLevel()
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			update[i] = sl.header
		}
		sl.level = level
	}

	newNode := newSortedSetNode(name, score, level)
	for i := 0; i < level; i++ {
		newNode.next[i] = update[i].next[i]
		update[i].next[i] = newNode
	}

	// Add the node to the hash table
	sl.table[name] = newNode
}

// Remove deletes an element by name from the skip list.
func (sl *SortedSet) Remove(name string) {
	node, exists := sl.table[name]
	if !exists {
		return
	}
	score := node.score

	update := make([]*sortedSetNode, MaxLevel)
	current := sl.header

	for i := sl.level - 1; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].score < score {
			current = current.next[i]
		}
		update[i] = current
	}

	current = current.next[0]

	if current != nil && current.name == name {
		for i := 0; i < sl.level; i++ {
			if update[i].next[i] != current {
				break
			}
			update[i].next[i] = current.next[i]
		}

		// Adjust the level of the skip list if necessary.
		for sl.level > 1 && sl.header.next[sl.level-1] == nil {
			sl.level--
		}
	}

	// Remove the node from the hash table
	delete(sl.table, name)
}

// GetScore returns the score of an element by name.
func (sl *SortedSet) GetScore(name string) (int, bool) {
	node, exists := sl.table[name]
	if !exists {
		return 0, false
	}
	return node.score, true
}
