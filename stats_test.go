package sampdo

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSampdoAppend(t *testing.T) {
	t.Run("valid points", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(1.5, 2.5, 3.0))
		require.Equal(t, []float64{1.5, 2.5, 3.0}, s.points)
		require.False(t, s.hasNegative)
	})

	t.Run("negative points", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(-1.5, 2.5))
		require.Equal(t, []float64{-1.5, 2.5}, s.points)
		require.True(t, s.hasNegative)
	})

	t.Run("rejects NaN", func(t *testing.T) {
		s := New()
		err := s.Append(math.NaN())
		require.Error(t, err)
	})

	t.Run("accepts negative zero using comparison sort", func(t *testing.T) {
		s := New()
		err := s.Append(math.Copysign(0, -1))
		require.NoError(t, err)
		require.True(t, s.hasNegative)
	})
}

func TestSampdoSorted(t *testing.T) {
	t.Run("sorts non-negative points with radix sort", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(3.0, 1.0, 2.0))
		sorted, err := s.Sorted()
		require.NoError(t, err)
		require.Equal(t, []float64{1.0, 2.0, 3.0}, sorted.points)
	})

	t.Run("sorts negative points with slice sort", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(-3.0, -1.0, -2.0))
		sorted, err := s.Sorted()
		require.NoError(t, err)
		require.Equal(t, []float64{-3.0, -2.0, -1.0}, sorted.points)
	})

	t.Run("nil receiver", func(t *testing.T) {
		var s *Sampdo
		_, err := s.Sorted()
		require.Error(t, err)
	})

	t.Run("empty points", func(t *testing.T) {
		s := New()
		_, err := s.Sorted()
		require.Error(t, err)
	})
}

func TestSampdoSummary(t *testing.T) {
	tests := []struct {
		name    string
		floats  []float64
		ptSet   []float64
		wantMin float64
		wantMax float64
		wantAvg float64
		wantPts []float64
		wantErr bool
	}{
		{
			name:    "basic values with default percentiles",
			floats:  []float64{8.0, 2.0, 10.0, 4.0, 6.0, 3.0, 7.0, 1.0, 9.0, 5.0},
			ptSet:   []float64{99, 95, 90, 75},
			wantMin: 1,
			wantMax: 10,
			wantAvg: 5.5,
			wantPts: []float64{9.91, 9.549999999999999, 9.1, 7.75},
		},
		{
			name:    "single value",
			floats:  []float64{42.0},
			ptSet:   []float64{50},
			wantMin: 42,
			wantMax: 42,
			wantAvg: 42,
			wantPts: []float64{42},
		},
		{
			name:    "two values",
			floats:  []float64{1.0, 3.0},
			ptSet:   []float64{50},
			wantMin: 1,
			wantMax: 3,
			wantAvg: 2,
			wantPts: []float64{2},
		},
		{
			name:    "empty percentile set",
			floats:  []float64{1.0, 2.0, 3.0},
			ptSet:   []float64{},
			wantMin: 1,
			wantMax: 3,
			wantAvg: 2,
			wantPts: []float64{},
		},
		{
			name:    "unsorted input",
			floats:  []float64{5.0, 1.0, 3.0, 2.0, 4.0},
			ptSet:   []float64{0, 100, 50},
			wantMin: 1,
			wantMax: 5,
			wantAvg: 3,
			wantPts: []float64{1, 5, 3},
		},
		{
			name:    "decimal percentiles",
			floats:  []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			ptSet:   []float64{12.5, 87.5},
			wantMin: 1,
			wantMax: 5,
			wantAvg: 3,
			wantPts: []float64{1.5, 4.5},
		},
		{
			name:    "duplicate values",
			floats:  []float64{5.0, 5.0, 5.0, 5.0},
			ptSet:   []float64{25, 50, 75},
			wantMin: 5,
			wantMax: 5,
			wantAvg: 5,
			wantPts: []float64{5, 5, 5},
		},
		{
			name:    "negative values",
			floats:  []float64{-5.0, -1.0, -3.0},
			ptSet:   []float64{50},
			wantMin: -5,
			wantMax: -1,
			wantAvg: -3,
			wantPts: []float64{-3},
		},
		{
			name:    "empty input",
			floats:  []float64{},
			ptSet:   []float64{50},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := New()
			require.NoError(t, data.Append(tt.floats...))
			sorted, err := data.Sorted()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			min, err := sorted.Min()
			require.NoError(t, err)
			require.InDelta(t, tt.wantMin, min, 1e-9)

			max, err := sorted.Max()
			require.NoError(t, err)
			require.InDelta(t, tt.wantMax, max, 1e-9)

			avg, err := sorted.Mean()
			require.NoError(t, err)
			require.InDelta(t, tt.wantAvg, avg, 1e-9)
			for i, ps := range tt.ptSet {
				pt, err := sorted.Percentile(ps)
				require.NoError(t, err)
				require.InDelta(t, tt.wantPts[i], pt, 1e-9)
			}
		})
	}
}

func TestSampdoMeanPreservesInputOrder(t *testing.T) {
	data := New()
	require.NoError(t, data.Append(1e16, 1, -1e16))
	sorted, err := data.Sorted()
	require.NoError(t, err)
	mean, err := sorted.Mean()
	require.NoError(t, err)
	require.Equal(t, 0.0, mean)
	require.Error(t, data.Append(-2, math.NaN()))
	sorted, err = data.Sorted()
	require.NoError(t, err)
	require.Equal(t, 3, sorted.Count())
	mean, err = sorted.Mean()
	require.NoError(t, err)
	require.Equal(t, 0.0, mean)
}

func TestSampdoFrozenAppend(t *testing.T) {
	s := New()
	require.NoError(t, s.Append(3.0, 1.0, 2.0))
	sorted, err := s.Sorted()
	require.NoError(t, err)
	require.Equal(t, []float64{1.0, 2.0, 3.0}, sorted.points)

	// Appending after Sorted must not mutate the already-sorted slice.
	require.NoError(t, s.Append(0.5))
	require.Equal(t, []float64{1.0, 2.0, 3.0}, sorted.points)

	sorted2, err := s.Sorted()
	require.NoError(t, err)
	require.Equal(t, []float64{0.5, 1.0, 2.0, 3.0}, sorted2.points)
}

func TestSampdoCount(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		s := New()
		require.Equal(t, 0, s.Count())
	})

	t.Run("after append", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(1.0, 2.0, 3.0))
		require.Equal(t, 3, s.Count())
	})

	t.Run("after sorted", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(1.0, 2.0, 3.0))
		sorted, err := s.Sorted()
		require.NoError(t, err)
		require.Equal(t, 3, sorted.Count())
	})
}

func TestSampdoCopyTo(t *testing.T) {
	t.Run("copies points", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(1.0, 2.0, 3.0))
		dst := make([]float64, 0, 3)
		err := s.CopyTo(&dst)
		require.NoError(t, err)
		require.Equal(t, []float64{1.0, 2.0, 3.0}, dst)
	})

	t.Run("appends to existing slice", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(4.0, 5.0))
		dst := []float64{1.0, 2.0, 3.0}
		err := s.CopyTo(&dst)
		require.NoError(t, err)
		require.Equal(t, []float64{1.0, 2.0, 3.0, 4.0, 5.0}, dst)
	})

	t.Run("nil destination", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(1.0))
		err := s.CopyTo(nil)
		require.Error(t, err)
	})

	t.Run("nil underlying slice", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(1.0))
		var dst []float64
		err := s.CopyTo(&dst)
		require.Error(t, err)
	})
}

func TestSampdoAppendTo(t *testing.T) {
	t.Run("appends points", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(1.0, 2.0, 3.0))
		dst := New()
		require.NoError(t, s.AppendTo(dst))
		require.Equal(t, []float64{1.0, 2.0, 3.0}, dst.points)
	})

	t.Run("appends to existing points", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(4.0, 5.0))
		dst := New()
		require.NoError(t, dst.Append(1.0, 2.0, 3.0))
		require.NoError(t, s.AppendTo(dst))
		require.Equal(t, []float64{1.0, 2.0, 3.0, 4.0, 5.0}, dst.points)
		require.Equal(t, 15.0, dst.sum)
	})

	t.Run("nil destination", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(1.0))
		err := s.AppendTo(nil)
		require.Error(t, err)
	})

	t.Run("preserves sum", func(t *testing.T) {
		s := New()
		require.NoError(t, s.Append(1.0, 2.0, 3.0))
		dst := New()
		require.NoError(t, s.AppendTo(dst))
		require.Equal(t, 6.0, dst.sum)
	})
}
