package sampdo

import (
	"errors"
	"math"
)

var ErrNoPoints = errors.New("no points available")
var ErrInvalidPercentile = errors.New("invalid percentile")

// ErrInsufficientPoints indicates that sample standard deviation needs at least two points.
var ErrInsufficientPoints = errors.New("at least two points required")

type Sorted struct {
	points   []float64
	sum      float64
	sumValid bool
}

func (s *Sorted) Count() int {
	return len(s.points)
}

func (s *Sorted) Max() (float64, error) {
	if len(s.points) == 0 {
		return 0, ErrNoPoints
	}
	return s.points[len(s.points)-1], nil
}

func (s *Sorted) Min() (float64, error) {
	if len(s.points) == 0 {
		return 0, ErrNoPoints
	}
	return s.points[0], nil
}

func (s *Sorted) Mean() (float64, error) {
	if len(s.points) == 0 {
		return 0, ErrNoPoints
	}
	if s.sumValid {
		return s.sum / float64(len(s.points)), nil
	}
	sum := 0.0
	for _, v := range s.points {
		sum += v
	}
	return sum / float64(len(s.points)), nil
}

// PopulationStdDev returns the population standard deviation (divisor n).
// Empty input returns 0 and ErrNoPoints. Input containing infinity returns NaN.
func (s *Sorted) PopulationStdDev() (float64, error) {
	return s.stdDev(0)
}

// SampleStdDev returns the sample standard deviation (divisor n-1).
// Empty input returns 0 and ErrNoPoints; one point returns 0 and
// ErrInsufficientPoints. Otherwise, input containing infinity returns NaN.
func (s *Sorted) SampleStdDev() (float64, error) {
	return s.stdDev(1)
}

func (s *Sorted) stdDev(correction int) (float64, error) {
	n := len(s.points)
	if n == 0 {
		return 0, ErrNoPoints
	}
	if n <= correction {
		return 0, ErrInsufficientPoints
	}
	low, high := s.points[0], s.points[n-1]
	if math.IsInf(low, 0) || math.IsInf(high, 0) {
		return math.NaN(), nil
	}
	if low == high {
		return 0, nil
	}
	// Center before scaling to preserve small differences on a large offset.
	// Halving the endpoints first avoids overflow for extreme finite inputs.
	center := low/2 + high/2
	scale := math.Max(math.Abs(low-center), math.Abs(high-center))
	// Compensate the sum of normalized offsets; the cached input-order sum
	// can overflow or lose the very differences needed by this calculation.
	sum, compensation := 0.0, 0.0
	for _, v := range s.points {
		y := (v-center)/scale - compensation
		next := sum + y
		compensation = (next - sum) - y
		sum = next
	}
	mean := sum / float64(n)
	squares, compensation := 0.0, 0.0
	for _, v := range s.points {
		d := (v-center)/scale - mean
		y := d*d - compensation
		next := squares + y
		compensation = (next - squares) - y
		squares = next
	}
	return scale * math.Sqrt(squares/float64(n-correction)), nil
}

func (s *Sorted) Median() (float64, error) {
	if len(s.points) == 0 {
		return 0, ErrNoPoints
	}
	mid := len(s.points) / 2
	if len(s.points)%2 == 0 {
		if math.IsInf(s.points[mid-1], 0) || math.IsInf(s.points[mid], 0) {
			return s.Percentile(50)
		}
		return (s.points[mid-1] + s.points[mid]) / 2, nil
	}
	return s.points[mid], nil
}

func (s *Sorted) Percentile(p float64) (float64, error) {
	if len(s.points) == 0 {
		return 0, ErrNoPoints
	}
	if math.IsNaN(p) || p < 0 || p > 100 {
		return math.NaN(), ErrInvalidPercentile
	}
	k := (p / 100) * float64(len(s.points)-1)
	f := int(k)
	c := f + 1
	if c >= len(s.points) {
		return s.points[f], nil
	}
	fraction := k - float64(f)
	a, b := s.points[f], s.points[c]
	if fraction == 0 || a == b {
		return a, nil
	}
	if math.IsInf(a, -1) && math.IsInf(b, 1) {
		return math.NaN(), nil
	}
	if math.IsInf(a, 0) {
		return a, nil
	}
	if math.IsInf(b, 0) {
		return b, nil
	}
	return a + (b-a)*fraction, nil
}
