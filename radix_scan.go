package sampdo

import "math"

// radixScanGeneric is the portable reference for the SIMD scan. Non-negative,
// non-NaN IEEE 754 bit patterns have the same ordering as their numeric values.
func radixScanGeneric(points []float64) (varying uint64, ordered bool) {
	if len(points) == 0 {
		return 0, true
	}
	first := math.Float64bits(points[0])
	previous := first
	ordered = true
	for _, v := range points {
		key := math.Float64bits(v)
		varying |= key ^ first
		ordered = ordered && previous <= key
		previous = key
	}
	return
}
