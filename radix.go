package sampdo

import (
	"math"
	"math/bits"
	"slices"
)

// radixSort accepts non-negative float64 values, excluding NaN and negative zero.
// Sampdo.Append and Sampdo.Sorted enforce these preconditions. The returned slice
// always shares the input's backing array, including an odd number of radix passes.
func radixSort(points []float64) []float64 {
	if len(points) < 512 {
		slices.Sort(points)
		return points
	}
	// A descending input only needs a reversal. Check the endpoints first so
	// ascending/equal inputs avoid this scan; unordered inputs exit it early.
	if points[0] > points[len(points)-1] && radixDescending(points) {
		slices.Reverse(points)
		return points
	}
	varying, ordered := radixScan(points)
	if ordered {
		return points
	}
	scratch := make([]float64, len(points))
	// Small partitions amortize the compact MSD histogram. Repeated full-width
	// keys also benefit: equal buckets finish without visiting the remaining bits.
	// Narrow keys need few LSD passes, so don't sample them.
	if len(points) < 2048 || bits.Len64(varying)-bits.TrailingZeros64(varying) > 22 && radixRepeatedSample(points) {
		radixSortMSDInto(points, scratch, varying)
	} else {
		radixSortLSD(points, scratch, varying)
	}
	return points
}

// Equal neighbors are allowed. Confirm the entire input before reversing it:
// a descending prefix or descending endpoints alone are not sufficient.
func radixDescending(points []float64) bool {
	for i := 1; i < len(points); i++ {
		if points[i] > points[i-1] {
			return false
		}
	}
	return true
}

// Sample 512 evenly spaced keys without allocating (len(points) >= 2048).
// This only chooses
// an algorithm; neither correctness nor the worst-case radix depth depends on
// the sample being representative. Stop once at least 1/8 of the sample repeats.
func radixRepeatedSample(points []float64) bool {
	var seen [1024]uint64
	repeats := 0
	step := len(points) / 512
	for i := range 512 {
		key := math.Float64bits(points[i*step]) + 1 // zero is the empty slot sentinel
		slot := key * 0x9e3779b97f4a7c15 >> 54
		for seen[slot] != 0 && seen[slot] != key {
			slot = (slot + 1) & 1023
		}
		if seen[slot] == key {
			repeats++
			if repeats == 64 {
				return true
			}
		}
		seen[slot] = key
	}
	return false
}

// Build all six histograms in one traversal. Eleven bits reduce the number of
// scatter passes from eight to six while keeping each histogram at 16 KiB on
// 64-bit systems. Histograms are independent of the order of previous passes.
func radixSortLSD(points, scratch []float64, varying uint64) {
	var counts [6][2048]int
	for _, v := range points {
		k := math.Float64bits(v)
		counts[0][k&2047]++
		counts[1][(k>>11)&2047]++
		counts[2][(k>>22)&2047]++
		counts[3][(k>>33)&2047]++
		counts[4][(k>>44)&2047]++
		counts[5][k>>55]++
	}
	src, dst := points, scratch
	for pass := range 6 {
		shift := uint(pass * 11)
		if (varying>>shift)&2047 == 0 {
			continue
		}
		c := &counts[pass]
		sum := 0
		for i, n := range c {
			c[i] = sum
			sum += n
		}
		for _, v := range src {
			idx := (math.Float64bits(v) >> shift) & 2047
			dst[c[idx]] = v
			c[idx]++
		}
		src, dst = dst, src
	}
	if &src[0] != &points[0] {
		copy(points, src)
	}
}

// Partition from the most significant varying byte. Recursion skips shared
// prefixes and stops on ordered/equal buckets; at most eight levels are needed.
func radixSortMSDInto(points, scratch []float64, v uint64) {
	shift := max(0, bits.Len64(v)-8)
	count := radixPartition(points, scratch, uint(shift))
	copy(points, scratch)
	if shift == 0 {
		return
	}
	start := 0
	for _, end := range count {
		if end-start > 128 {
			vary, sorted := radixScan(points[start:end])
			if !sorted {
				radixSortMSDInto(points[start:end], scratch[start:end], vary)
			}
		} else if end-start > 1 {
			slices.Sort(points[start:end])
		}
		start = end
	}
}

// Four independent counters break the histogram dependency chain when many
// adjacent keys share a byte. After scattering, offsets are bucket ends.
func radixPartition(points, scratch []float64, shift uint) [256]int {
	var counts [4][256]int
	i := 0
	for ; i+4 <= len(points); i += 4 {
		counts[0][(math.Float64bits(points[i])>>shift)&255]++
		counts[1][(math.Float64bits(points[i+1])>>shift)&255]++
		counts[2][(math.Float64bits(points[i+2])>>shift)&255]++
		counts[3][(math.Float64bits(points[i+3])>>shift)&255]++
	}
	for ; i < len(points); i++ {
		counts[0][(math.Float64bits(points[i])>>shift)&255]++
	}
	count := &counts[0]
	sum := 0
	for i := range count {
		n := counts[0][i] + counts[1][i] + counts[2][i] + counts[3][i]
		count[i] = sum
		sum += n
	}
	for _, val := range points {
		idx := (math.Float64bits(val) >> shift) & 255
		scratch[count[idx]] = val
		count[idx]++
	}
	return *count
}
