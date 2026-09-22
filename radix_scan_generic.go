//go:build (!arm64 && !amd64) || purego || race

package sampdo

func radixScan(points []float64) (uint64, bool) {
	return radixScanGeneric(points)
}
