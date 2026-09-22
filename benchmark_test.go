package sampdo

import (
	"fmt"
	"testing"
)

func sampdoAll(b *testing.B, distribution string) float64 {
	ptiles := []float64{75, 90, 95, 99}
	b.StopTimer()
	points := radixInput(100_000, distribution)
	sink := 0.0
	b.StartTimer()

	s := New()
	err := s.Append(points...)
	if err != nil {
		b.Fatal(err)
	}
	sorted, err := s.Sorted()
	if err != nil {
		b.Fatal(err)
	}

	max, _ := sorted.Max()
	sink += max
	min, _ := sorted.Min()
	sink += min
	mean, _ := sorted.Mean()
	sink += mean
	median, _ := sorted.Median()
	sink += median

	for _, p := range ptiles {
		val, _ := sorted.Percentile(p)
		sink += val
	}
	return sink
}

func BenchmarkSampdoAll(b *testing.B) {
	for _, distribution := range []string{"duplicates", "random", "wide", "sorted", "response_time"} {
		b.Run(fmt.Sprintf("sampdo/%d/%s", 100_000, distribution), func(b *testing.B) {
			for b.Loop() {
				sampdoAll(b, distribution)
			}
		})
	}
}

func BenchmarkSortDistributions(b *testing.B) {
	for _, algorithm := range []struct {
		name string
		sort func([]float64) []float64
	}{
		{"slices", aliasSlicesSort}, {"radix", radixSort},
	} {
		for _, n := range []int{512, 1000, 2048, 4096, 10_000, 100_000, 1_000_000} {
			for _, distribution := range []string{"duplicates", "random", "wide", "sorted", "reverse", "equal", "response_time"} {
				{
					b.Run(fmt.Sprintf("%s/%d/%s", algorithm.name, n, distribution), func(b *testing.B) {
						input := radixInput(n, distribution)
						points := make([]float64, n)
						b.ReportAllocs()
						b.ResetTimer()
						// Include the same reset copy for both algorithms. Reuse
						// storage so input generation/GC does not distort timing.
						for b.Loop() {
							copy(points, input)
							algorithm.sort(points)
						}
					})
				}
			}
		}
	}
}

func BenchmarkRadixScan(b *testing.B) {
	for _, n := range []int{1000, 100_000} {
		for _, distribution := range []string{"wide", "sorted"} {
			points := radixInput(n, distribution)
			for _, scan := range []struct {
				name string
				scan func([]float64) (uint64, bool)
			}{{"native", radixScan}, {"generic", radixScanGeneric}} {
				b.Run(fmt.Sprintf("%s/%d/%s", scan.name, n, distribution), func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(int64(n) * 8)
					for b.Loop() {
						scan.scan(points)
					}
				})
			}
		}
	}
}
