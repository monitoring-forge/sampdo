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
		for _, n := range []int{1000, 10_000, 100_000} {
			for _, distribution := range []string{"duplicates", "random", "wide", "sorted", "response_time"} {
				{
					b.Run(fmt.Sprintf("%s/%d/%s", algorithm.name, n, distribution), func(b *testing.B) {
						b.ReportAllocs()
						for b.Loop() {
							b.StopTimer()
							points := radixInput(n, distribution)
							b.StartTimer()
							algorithm.sort(points)
						}
					})
				}
			}
		}
	}
}
