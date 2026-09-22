//go:build !purego && !race

package sampdo

// radixScan uses SSE2, available on every amd64 CPU. No AVX feature check is needed.
//
//go:noescape
func radixScan(points []float64) (varying uint64, ordered bool)
