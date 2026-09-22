//go:build !purego && !race

package sampdo

// radixScan uses baseline ARM64 NEON instructions. It reads only within points.
//
//go:noescape
func radixScan(points []float64) (varying uint64, ordered bool)
