package datatypes

import (
	"hash"
	"hash/fnv"
	"math"
)

type BloomFilter struct {
	bitSet        []bool
	numHashes     uint
	bitArraySize  uint
	expectedItems uint
	hashes        []hash.Hash64
}

// NewBloomFilter creates a new Bloom filter with the specified number of expected items (expectedItems)
// and false positive probability (p).
func NewBloomFilter(expectedItems uint, errorRate float64) *BloomFilter {
	bitArraySize := optimalM(expectedItems, errorRate)
	numHashes := optimalK(expectedItems, bitArraySize)
	hashes := make([]hash.Hash64, numHashes)

	// Initialize hash functions
	for i := range hashes {
		hashes[i] = fnv.New64a() // Using FNV-1a hash function
	}

	return &BloomFilter{
		bitSet:        make([]bool, bitArraySize),
		numHashes:     numHashes,
		bitArraySize:  bitArraySize,
		expectedItems: expectedItems,
		hashes:        hashes,
	}
}

// optimalM calculates the optimal size of the bit array (bitArraySize) given the expected
// number of items (expectedItems) and the desired false positive probability (p).
func optimalM(n uint, p float64) uint {
	return uint(math.Ceil(float64(n) * math.Log(p) / math.Log(1/math.Pow(2, math.Log(2)))))
}

// optimalK calculates the optimal number of hash functions (numHashes) given the size
// of the bit array (bitArraySize) and the expected number of items (expectedItems).
func optimalK(n, m uint) uint {
	return uint(math.Round(float64(m) / float64(n) * math.Log(2)))
}

// Add inserts an item into the Bloom filter.
func (bf *BloomFilter) Add(item string) {
	for _, h := range bf.hashes {
		bf.setBit(h, item)
	}
}

// Contains checks if an item is possibly in the Bloom filter.
// Returns true if the item is possibly in the set, false if it is definitely not in the set.
func (bf *BloomFilter) Contains(item string) bool {
	for _, h := range bf.hashes {
		if !bf.getBit(h, item) {
			return false
		}
	}
	return true
}

// setBit sets the bit corresponding to the hashed value of the item.
func (bf *BloomFilter) setBit(h hash.Hash64, item string) {
	index := bf.indexFor(h, item)
	bf.bitSet[index] = true
}

// getBit checks if the bit corresponding to the hashed value of the item is set.
func (bf *BloomFilter) getBit(h hash.Hash64, item string) bool {
	index := bf.indexFor(h, item)
	return bf.bitSet[index]
}

// indexFor calculates the index in the bit array for the given hash function and item.
func (bf *BloomFilter) indexFor(h hash.Hash64, item string) uint {
	h.Reset()
	h.Write([]byte(item))
	hashValue := h.Sum64()
	return uint(hashValue % uint64(bf.bitArraySize))
}
