package datatypes

import (
	"encoding/json"
	"hash/fnv"
	"math"
	"math/bits"
)

// HyperLogLog represents the HLL data structure
type HyperLogLog struct {
	p         uint8
	m         uint32
	registers []uint8
	alphaMM   float64
	Name      string `json:"name"`
}

func (hll *HyperLogLog) MarshalJSON() ([]byte, error) {
	return json.Marshal(hll)
}

func (hll *HyperLogLog) GetName() string {
	return hll.Name
}

// NewHllWithErrorRate creates a new HyperLogLog with the specified error rate as a percentage
func NewHllWithErrorRate(errorRate float64) *HyperLogLog {
	if errorRate <= 0.05 || errorRate >= 100 {
		panic("Error rate must be between 0.5 and 100 (exclusive)")
	}

	// Convert error rate percentage to a decimal
	errorRate = errorRate / 100.0

	// Calculate precision p
	p := calculatePrecision(errorRate)

	return New(p)
}

// calculatePrecision calculates the precision p from the given error rate
func calculatePrecision(errorRate float64) uint8 {
	// Formula: p = log2((1.04 / errorRate) ^ 2)
	p := math.Log2(math.Pow(1.04/errorRate, 2))
	return uint8(math.Ceil(p)) // Round up to the nearest integer
}

// New creates a new HyperLogLog with the specified precision p
func New(p uint8) *HyperLogLog {
	m := uint32(1) << p
	registers := make([]uint8, m)

	alphaMM := getAlphaMM(m)

	return &HyperLogLog{
		p:         p,
		m:         m,
		registers: registers,
		alphaMM:   alphaMM,
		Name:      "hll",
	}
}

// Add adds an element to the HyperLogLog
func (hll *HyperLogLog) Add(item string) {
	hash := hash64(item)

	// Get the first p bits as the register index
	registerIndex := hash >> (64 - hll.p)

	// Get the rank (leading zeroes of the remaining bits + 1)
	rank := bits.LeadingZeros64((hash<<hll.p)|(1<<(hll.p-1))) + 1

	// Update the register if the rank is higher
	if uint8(rank) > hll.registers[registerIndex] {
		hll.registers[registerIndex] = uint8(rank)
	}
}

// AddMany adds an element to the HyperLogLog
func (hll *HyperLogLog) AddMany(items []string) int {
	oldCardinality := hll.Estimate()
	for _, item := range items {
		hll.Add(item)
	}
	newCardinality := hll.Estimate()
	if oldCardinality != newCardinality {
		return 1
	}
	return 0
}

// Estimate returns the estimated cardinality
func (hll *HyperLogLog) Estimate() float64 {
	var sum float64

	for _, reg := range hll.registers {
		sum += 1.0 / math.Pow(2.0, float64(reg))
	}

	estimate := hll.alphaMM / sum

	// Apply corrections for small and large cardinalities
	if estimate <= 2.5*float64(hll.m) {
		v := float64(hll.countZeroRegisters())
		if v > 0 {
			estimate = float64(hll.m) * math.Log(float64(hll.m)/v)
		}
	} else if estimate > (1.0/30.0)*math.Pow(2.0, 64.0) {
		estimate = -math.Pow(2.0, 64.0) * math.Log(1.0-estimate/math.Pow(2.0, 64.0))
	}

	return math.Round(estimate)
}

// MergeArrayIntoNew merges an array of HyperLogLog instances into a new one
// without modifying the originals
func MergeArrayIntoNew(hlls []*HyperLogLog) *HyperLogLog {
	if len(hlls) == 0 {
		panic("No HyperLogLogs provided for merging")
	}

	// Ensure all HyperLogLogs have the same precision
	precision := hlls[0].p
	for _, hll := range hlls {
		if hll.p != precision {
			panic("HyperLogLog precision mismatch")
		}
	}

	// Create a new HyperLogLog with the same precision
	newHLL := New(precision)

	// Merge all the HyperLogLog registers into the new HyperLogLog
	for _, hll := range hlls {
		for i := range newHLL.registers {
			if hll.registers[i] > newHLL.registers[i] {
				newHLL.registers[i] = hll.registers[i]
			}
		}
	}

	return newHLL
}

// countZeroRegisters counts the number of registers that are still zero
func (hll *HyperLogLog) countZeroRegisters() int {
	count := 0
	for _, reg := range hll.registers {
		if reg == 0 {
			count++
		}
	}
	return count
}

// hash64 returns a 64-bit FNV-1a hash of a string
func hash64(item string) uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(item))
	return hasher.Sum64()
}

// getAlphaMM computes the alphaMM constant for a given m
func getAlphaMM(m uint32) float64 {
	switch m {
	case 16:
		return 0.673 * float64(m*m)
	case 32:
		return 0.697 * float64(m*m)
	case 64:
		return 0.709 * float64(m*m)
	default:
		return (0.7213 / (1.0 + 1.079/float64(m))) * float64(m*m)
	}
}
