package datatypes

import (
	"testing"
)

func TestBloomFilter(t *testing.T) {
	bf := NewBloomFilter(100, 0.01)

	bf.Add("Hello")
	bf.Add("World")

	found := bf.Contains("Something")
	if found {
		t.Errorf("Bloom filter should not contain \"Something\"")
	}
}
