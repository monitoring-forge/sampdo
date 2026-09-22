package sampdo

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRadixScan(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 7, 8, 9, 15, 16, 17, 31, 32, 33, 127, 128, 129, 511, 512, 513, 2047, 2048, 2049} {
		for _, d := range []string{"random", "wide", "sorted", "reverse", "equal", "response_time"} {
			// Offset the slice to exercise unaligned SIMD loads as well as tails.
			storage := make([]float64, n+2)
			points := storage[1 : n+1]
			copy(points, radixInput(n, d))
			wantV, wantO := radixScanGeneric(points)
			gotV, gotO := radixScan(points)
			require.Equal(t, wantV, gotV, "varying %s/%d", d, n)
			require.Equal(t, wantO, gotO, "ordered %s/%d", d, n)
		}
	}
}

func TestRadixScanEveryInversion(t *testing.T) {
	// Check both SIMD lanes, vector boundaries, and the scalar tail, including
	// an inversion that is the only difference from an otherwise ordered slice.
	for n := 2; n <= 65; n++ {
		for i := 0; i < n-1; i++ {
			points := radixInput(n, "sorted")
			points[i], points[i+1] = points[i+1], points[i]
			want, _ := radixScanGeneric(points)
			got, ordered := radixScan(points)
			require.Equal(t, want, got, "n=%d inversion=%d", n, i)
			require.False(t, ordered, "n=%d inversion=%d", n, i)
		}
	}
	for bit := range 63 {
		for i := range 33 {
			points := make([]float64, 33)
			points[i] = math.Float64frombits(uint64(1) << bit)
			got, ordered := radixScan(points)
			require.Equal(t, uint64(1)<<bit, got)
			require.Equal(t, i == 32, ordered)
		}
	}
}

func checkRadixInPlace(t *testing.T, input []float64) {
	t.Helper()
	want := slices.Clone(input)
	slices.Sort(want)
	storage := make([]float64, len(input)+2)
	storage[0], storage[len(storage)-1] = -123, -456
	points := storage[1 : len(storage)-1]
	copy(points, input)
	got := radixSort(points)
	require.Equal(t, want, got)
	require.Equal(t, want, points, "input must be sorted even if return value is ignored")
	if len(points) > 0 {
		require.Same(t, &points[0], &got[0])
	}
	require.Equal(t, -123.0, storage[0])
	require.Equal(t, -456.0, storage[len(storage)-1])
}

func TestRadixSortInPlace(t *testing.T) {
	for _, n := range []int{511, 512, 513, 1023, 2047, 2048, 2049, 8193, 100000} {
		for _, d := range []string{"duplicates", "random", "wide", "sorted", "reverse", "equal", "response_time", "nearly_sorted"} {
			t.Run(fmt.Sprintf("%s/%d", d, n), func(t *testing.T) { checkRadixInPlace(t, radixInput(n, d)) })
		}
	}
	// Force every LSD digit individually, and odd/even numbers of active digits.
	// These also exercise all-subnormal inputs and skipping common middle digits.
	for passes := 1; passes <= 6; passes++ {
		for start := 0; start+passes <= 6; start++ {
			points := make([]float64, 4097)
			for i := range points {
				var key uint64
				for p := start; p < start+passes; p++ {
					key |= uint64((i*17+p)%8) << uint(p*11)
				}
				points[i] = math.Float64frombits(key)
			}
			checkRadixInPlace(t, points)
		}
	}
}

func TestRadixSortPaths(t *testing.T) {
	for _, d := range []string{"duplicates", "random", "wide", "response_time", "reverse", "equal"} {
		original := radixInput(8193, d)
		want := slices.Clone(original)
		slices.Sort(want)
		varying, _ := radixScanGeneric(original)
		for _, sort := range []func([]float64, []float64, uint64){radixSortLSD, radixSortMSDInto} {
			points := slices.Clone(original)
			sort(points, make([]float64, len(points)), varying)
			require.Equal(t, want, points, d)
		}
	}
	require.True(t, radixRepeatedSample(radixInput(10000, "response_time")))
	require.False(t, radixRepeatedSample(radixInput(10000, "wide")))
}

func TestRadixSortBitPatterns(t *testing.T) {
	r := rand.New(rand.NewPCG(19, 37))
	for trial := range 30 {
		points := make([]float64, 2048+trial)
		mask := r.Uint64() & 0x7fefffffffffffff
		for i := range points {
			points[i] = math.Float64frombits(r.Uint64() & mask)
		}
		points[len(points)/2] = math.Inf(1)
		checkRadixInPlace(t, points)
	}
}

func FuzzRadixSort(f *testing.F) {
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0xf0, 0x7f, 1, 0, 0, 0, 0, 0, 0, 0})
	f.Add([]byte{1, 7, 2, 11, 3, 19, 5, 23, 8})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		n := min(4097, max(513, len(data)))
		points := make([]float64, n)
		for i := range points {
			var raw [8]byte
			for j := range raw {
				raw[j] = data[(i*8+j)%len(data)]
			}
			key := binary.LittleEndian.Uint64(raw[:]) & 0x7fffffffffffffff
			key = min(key, 0x7ff0000000000000)
			points[i] = math.Float64frombits(key)
		}
		checkRadixInPlace(t, points)
	})
}

func TestRadixSortOrderedAllocs(t *testing.T) {
	for _, distribution := range []string{"sorted", "equal"} {
		points := radixInput(8193, distribution)
		require.Zero(t, testing.AllocsPerRun(10, func() { radixSort(points) }), distribution)
	}
}
