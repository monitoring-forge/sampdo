package sampdo

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortedBasic(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		s := &Sorted{points: []float64{1.0, 2.0, 3.0}}
		require.Equal(t, 3, s.Count())
		min, err := s.Min()
		require.NoError(t, err)
		require.InDelta(t, 1.0, min, 1e-9)
		max, err := s.Max()
		require.NoError(t, err)
		require.InDelta(t, 3.0, max, 1e-9)
		mean, err := s.Mean()
		require.NoError(t, err)
		require.InDelta(t, 2.0, mean, 1e-9)
	})

	t.Run("empty", func(t *testing.T) {
		s := &Sorted{points: []float64{}}
		require.Equal(t, 0, s.Count())
		_, err := s.Min()
		require.Error(t, err)
		_, err = s.Max()
		require.Error(t, err)
		_, err = s.Mean()
		require.Error(t, err)
	})

}

func TestSortedStdDev(t *testing.T) {
	tests := []struct {
		name               string
		points             []float64
		population, sample float64
	}{
		{"textbook", []float64{2, 4, 4, 4, 5, 5, 7, 9}, 2, math.Sqrt(32.0 / 7)},
		{"negative", []float64{-5, -3, -1}, math.Sqrt(8.0 / 3), 2},
		{"mixed", []float64{-1, 0, 1}, math.Sqrt(2.0 / 3), 1},
		{"constant", []float64{42, 42, 42}, 0, 0},
		{"zeros", []float64{math.Copysign(0, -1), 0}, 0, 0},
		{"large offset", []float64{1e16, 1e16 + 2, 1e16 + 4}, math.Sqrt(8.0 / 3), 2},
		{"large magnitude", []float64{-1e200, 1e200}, 1e200, math.Sqrt2 * 1e200},
		{"small magnitude", []float64{-1e-200, 1e-200}, 1e-200, math.Sqrt2 * 1e-200},
		{"overflowing sum", []float64{math.MaxFloat64, math.MaxFloat64}, 0, 0},
		{"extreme range", []float64{-math.MaxFloat64, math.MaxFloat64}, math.MaxFloat64, math.Inf(1)},
		{"subnormal", []float64{0, 2 * math.SmallestNonzeroFloat64}, math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64},
		{"positive infinity", []float64{1, math.Inf(1)}, math.NaN(), math.NaN()},
		{"negative infinity", []float64{math.Inf(-1), 1}, math.NaN(), math.NaN()},
		{"both infinities", []float64{math.Inf(-1), math.Inf(1)}, math.NaN(), math.NaN()},
		{"equal infinities", []float64{math.Inf(1), math.Inf(1)}, math.NaN(), math.NaN()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := New()
			require.NoError(t, data.Append(tt.points...))
			sorted, err := data.Sorted()
			require.NoError(t, err)
			for _, view := range []*Sorted{sorted, {points: slices.Clone(sorted.points)}} {
				population, err := view.PopulationStdDev()
				require.NoError(t, err)
				sample, err := view.SampleStdDev()
				require.NoError(t, err)
				for i, got := range []float64{population, sample} {
					want := []float64{tt.population, tt.sample}[i]
					switch {
					case math.IsNaN(want):
						require.True(t, math.IsNaN(got))
					case want == 0 || math.IsInf(want, 0) || math.Abs(want) < 1e-300:
						require.Equal(t, want, got)
					default:
						require.InEpsilon(t, want, got, 1e-14)
					}
				}
			}
		})
	}
}

func TestSortedStdDevSmallInputs(t *testing.T) {
	s := &Sorted{}
	got, err := s.PopulationStdDev()
	require.ErrorIs(t, err, ErrNoPoints)
	require.Zero(t, got)
	got, err = s.SampleStdDev()
	require.ErrorIs(t, err, ErrNoPoints)
	require.Zero(t, got)
	for _, point := range []float64{42, math.MaxFloat64, math.Inf(1)} {
		s = &Sorted{points: []float64{point}}
		got, err = s.PopulationStdDev()
		require.NoError(t, err)
		if math.IsInf(point, 0) {
			require.True(t, math.IsNaN(got))
		} else {
			require.Zero(t, got)
		}
		got, err = s.SampleStdDev()
		require.ErrorIs(t, err, ErrInsufficientPoints)
		require.Zero(t, got)
	}
}

func TestSortedStdDevFrozenView(t *testing.T) {
	data := New()
	require.NoError(t, data.Append(3, 1))
	before, err := data.Sorted()
	require.NoError(t, err)
	require.NoError(t, data.Append(5))
	after, err := data.Sorted()
	require.NoError(t, err)
	got, err := before.PopulationStdDev()
	require.NoError(t, err)
	require.Equal(t, 1.0, got)
	got, err = before.SampleStdDev()
	require.NoError(t, err)
	require.Equal(t, math.Sqrt2, got)
	got, err = after.SampleStdDev()
	require.NoError(t, err)
	require.Equal(t, 2.0, got)
}

func TestSortedMedian(t *testing.T) {
	tests := []struct {
		name string
		pts  []float64
		want float64
	}{
		{
			name: "odd",
			pts:  []float64{1.0, 2.0, 3.0},
			want: 2.0,
		},
		{
			name: "even",
			pts:  []float64{1.0, 2.0, 3.0, 4.0},
			want: 2.5,
		},
		{
			name: "single",
			pts:  []float64{5.0},
			want: 5.0,
		},
		{
			name: "empty",
			pts:  []float64{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sorted{points: tt.pts}
			median, err := s.Median()
			if len(tt.pts) == 0 {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.InDelta(t, tt.want, median, 1e-9)
		})
	}
}

func TestSortedPercentile(t *testing.T) {
	tests := []struct {
		name    string
		pts     []float64
		p       float64
		want    float64
		wantErr bool
	}{
		{
			name:    "0th percentile",
			pts:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			p:       0,
			want:    1.0,
			wantErr: false,
		},
		{
			name:    "100th percentile",
			pts:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			p:       100,
			want:    5.0,
			wantErr: false,
		},
		{
			name:    "50th percentile odd",
			pts:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			p:       50,
			want:    3.0,
			wantErr: false,
		},
		{
			name:    "50th percentile even",
			pts:     []float64{1.0, 2.0, 3.0, 4.0},
			p:       50,
			want:    2.5,
			wantErr: false,
		},
		{
			name:    "out of range negative",
			pts:     []float64{1.0, 2.0, 3.0},
			p:       -1,
			want:    0,
			wantErr: true,
		},
		{
			name:    "out of range positive",
			pts:     []float64{1.0, 2.0, 3.0},
			p:       101,
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty",
			pts:     []float64{},
			p:       50,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sorted{points: tt.pts}
			got, err := s.Percentile(tt.p)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.InDelta(t, tt.want, got, 1e-9)
		})
	}
}

func TestInfinityPercentiles(t *testing.T) {
	tests := []struct {
		input   []float64
		percent float64
		want    float64
	}{
		{[]float64{1, math.Inf(1)}, 0, 1},
		{[]float64{1, math.Inf(1)}, 50, math.Inf(1)},
		{[]float64{math.Inf(-1), 1}, 50, math.Inf(-1)},
		{[]float64{math.Inf(1), math.Inf(1)}, 50, math.Inf(1)},
		{[]float64{math.Inf(-1), math.Inf(-1)}, 50, math.Inf(-1)},
		{[]float64{math.Inf(-1), math.Inf(1)}, 50, math.NaN()},
		{[]float64{math.Inf(-1), math.Inf(1)}, 100, math.Inf(1)},
	}
	for _, tt := range tests {
		data := New()
		require.NoError(t, data.Append(tt.input...))
		sorted, err := data.Sorted()
		require.NoError(t, err)
		got, err := sorted.Percentile(tt.percent)
		require.NoError(t, err)
		require.True(t, got == tt.want || math.IsNaN(got) && math.IsNaN(tt.want), "got %v want %v", got, tt.want)
	}
}

func TestRadixSort(t *testing.T) {
	tests := []struct {
		name  string
		input []float64
		want  []float64
	}{
		{
			name:  "basic",
			input: []float64{3.0, 1.0, 2.0},
			want:  []float64{1.0, 2.0, 3.0},
		},
		{
			name:  "with decimals",
			input: []float64{3.14, 1.59, 2.65},
			want:  []float64{1.59, 2.65, 3.14},
		},
		{
			name:  "duplicates",
			input: []float64{5.0, 5.0, 1.0, 1.0},
			want:  []float64{1.0, 1.0, 5.0, 5.0},
		},
		{
			name:  "empty",
			input: []float64{},
			want:  []float64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := slices.Clone(tt.input)
			got := radixSort(input)
			require.Equal(t, tt.want, got)
		})
	}
}

func aliasSlicesSort(points []float64) []float64 {
	slices.Sort(points)
	return points
}

func TestRadixSortMatchesSlices(t *testing.T) {
	for _, n := range []int{0, 1, 127, 128, 129, 511, 512, 513, 1000, 10000} {
		for _, distribution := range []string{"duplicates", "random", "wide", "sorted", "equal"} {
			t.Run(fmt.Sprintf("%s/%d", distribution, n), func(t *testing.T) {
				points := radixInput(n, distribution)
				want := aliasSlicesSort(slices.Clone(points))
				got := radixSort(slices.Clone(points))
				require.Equal(t, want, got)
			})
		}
	}
}

func TestRadixSortExtremeValues(t *testing.T) {
	points := make([]float64, 1024)
	extremes := []float64{math.Inf(1), math.MaxFloat64, math.SmallestNonzeroFloat64, 0, 1}
	for i := range points {
		points[i] = extremes[i%len(extremes)]
	}
	checkRadixInPlace(t, points)
}

func radixInput(n int, distribution string) []float64 {
	r := rand.New(rand.NewPCG(1, 2))
	points := make([]float64, n)
	for i := range points {
		switch distribution {
		case "duplicates":
			points[i] = float64(i%100) + 0.5
		case "random":
			points[i] = r.Float64() * 10000
		case "wide":
			points[i] = math.Float64frombits(r.Uint64() & 0x7fefffffffffffff)
		case "reverse_duplicates":
			points[i] = float64((n - 1 - i) / 8)
		case "reverse":
			points[i] = float64(n - i)
		case "nearly_sorted":
			points[i] = float64(i ^ 1)
		case "sorted":
			points[i] = float64(i)
		case "equal":
			points[i] = 42
		case "response_time":
			points[i] = float64(r.IntN(500)) / 1000
		}
	}
	return points
}
