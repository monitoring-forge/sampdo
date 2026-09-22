package sampdo

import (
	"errors"
	"math"
)

var ErrNoPoints = errors.New("no points available")
var ErrInvalidPercentile = errors.New("invalid percentile")

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
