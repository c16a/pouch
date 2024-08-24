package datatypes

import "testing"

func TestHyperLogLogMergeArrayIntoNew(t *testing.T) {
	hll1 := NewHllWithErrorRate(0.6)
	hll2 := NewHllWithErrorRate(0.6)
	hll3 := NewHllWithErrorRate(0.6)

	// Add some elements to each HLL
	hll1.Add("foo")
	hll1.Add("bar")

	hll2.Add("baz")
	hll2.Add("qux")

	hll3.Add("quux")
	hll3.Add("corge")

	// Merge an array of HLLs into a new HyperLogLog
	mergedHLL := MergeArrayIntoNew([]*HyperLogLog{hll1, hll2, hll3})

	// Estimate the cardinality of the merged HLL
	estimateMerged := mergedHLL.Estimate()

	// We expect the estimate to be around 6 since we added six unique elements
	if estimateMerged < 5 || estimateMerged > 7 {
		t.Errorf("Expected estimate around 6, got %f", estimateMerged)
	}

	// Ensure the original HLLs are not modified
	if hll1.Estimate() != 2 {
		t.Errorf("Expected estimate of hll1 to be 2, got %f", hll1.Estimate())
	}
	if hll2.Estimate() != 2 {
		t.Errorf("Expected estimate of hll2 to be 2, got %f", hll2.Estimate())
	}
	if hll3.Estimate() != 2 {
		t.Errorf("Expected estimate of hll3 to be 2, got %f", hll3.Estimate())
	}
}
