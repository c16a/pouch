package datatypes

import (
	"hash/fnv"
	"math/rand"
)

const (
	bucketSize      = 4   // Number of entries per bucket
	maxKicks        = 500 // Maximum number of kicks to relocate an entry
	fingerprintSize = 1   // Size of the fingerprint in bytes
)

type CuckooFilter struct {
	buckets [][bucketSize][]byte
	size    uint // Number of elements in the filter
}

// NewCuckooFilter creates a new Cuckoo filter with the specified number of buckets.
func NewCuckooFilter(numBuckets uint) *CuckooFilter {
	return &CuckooFilter{
		buckets: make([][bucketSize][]byte, numBuckets),
		size:    0,
	}
}

// Insert adds an item to the Cuckoo filter.
func (cf *CuckooFilter) Insert(item string) bool {
	fp := cf.fingerprint(item)
	i1 := cf.hash(item)
	i2 := cf.alternateIndex(i1, fp)

	if cf.insertToBucket(fp, i1) || cf.insertToBucket(fp, i2) {
		cf.size++
		return true
	}

	// Perform cuckoo evictions
	i := i1
	for n := 0; n < maxKicks; n++ {
		j := rand.Intn(bucketSize)
		evictedFp := cf.buckets[i][j]
		cf.buckets[i][j] = fp
		fp = evictedFp
		i = cf.alternateIndex(i, fp)
		if cf.insertToBucket(fp, i) {
			cf.size++
			return true
		}
	}

	return false // Filter is likely full
}

// Lookup checks if an item is in the Cuckoo filter.
func (cf *CuckooFilter) Lookup(item string) bool {
	fp := cf.fingerprint(item)
	i1 := cf.hash(item)
	i2 := cf.alternateIndex(i1, fp)

	return cf.bucketContains(fp, i1) || cf.bucketContains(fp, i2)
}

// Delete removes an item from the Cuckoo filter.
func (cf *CuckooFilter) Delete(item string) bool {
	fp := cf.fingerprint(item)
	i1 := cf.hash(item)
	i2 := cf.alternateIndex(i1, fp)

	if cf.deleteFromBucket(fp, i1) || cf.deleteFromBucket(fp, i2) {
		cf.size--
		return true
	}
	return false
}

// Size returns the number of items in the filter.
func (cf *CuckooFilter) Size() uint {
	return cf.size
}

func (cf *CuckooFilter) fingerprint(item string) []byte {
	h := fnv.New64a()
	h.Write([]byte(item))
	fp := make([]byte, fingerprintSize)
	copy(fp, h.Sum(nil)[:fingerprintSize])
	return fp
}

func (cf *CuckooFilter) hash(item string) uint {
	h := fnv.New64a()
	h.Write([]byte(item))
	return uint(h.Sum64() % uint64(len(cf.buckets)))
}

func (cf *CuckooFilter) alternateIndex(i uint, fp []byte) uint {
	h := fnv.New64a()
	h.Write(fp)
	return uint(i) ^ uint(h.Sum64()%uint64(len(cf.buckets)))
}

func (cf *CuckooFilter) insertToBucket(fp []byte, i uint) bool {
	for j := 0; j < bucketSize; j++ {
		if len(cf.buckets[i][j]) == 0 {
			cf.buckets[i][j] = fp
			return true
		}
	}
	return false
}

func (cf *CuckooFilter) bucketContains(fp []byte, i uint) bool {
	for j := 0; j < bucketSize; j++ {
		if string(cf.buckets[i][j]) == string(fp) {
			return true
		}
	}
	return false
}

func (cf *CuckooFilter) deleteFromBucket(fp []byte, i uint) bool {
	for j := 0; j < bucketSize; j++ {
		if string(cf.buckets[i][j]) == string(fp) {
			cf.buckets[i][j] = nil
			return true
		}
	}
	return false
}
