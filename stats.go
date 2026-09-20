package sampdo

import (
	"fmt"
	"math"
	"slices"
)

type Sampdo struct {
	PreferSlicesSort bool
	InitialCapacity  int
	points           []float64
	hasNegative      bool
	sum              float64
	frozen           bool
}

type Option func(*Sampdo)

func WithInitialCapacity(cap int) Option {
	return func(s *Sampdo) {
		s.InitialCapacity = cap
	}
}

func WithPreferSlicesSort(prefer bool) Option {
	return func(s *Sampdo) {
		s.PreferSlicesSort = prefer
	}
}

func New(opts ...Option) *Sampdo {
	s := &Sampdo{
		InitialCapacity: 128,
		hasNegative:     false,
		sum:             0,
	}
	for _, opt := range opts {
		opt(s)
	}
	if s.InitialCapacity > 0 {
		s.points = make([]float64, 0, s.InitialCapacity)
	}
	return s
}

func (t *Sampdo) Append(point ...float64) error {
	// Keep batch updates atomic while summing in the original input order.
	sum, hasNegative := t.sum, t.hasNegative
	for _, p := range point {
		if math.IsNaN(p) {
			return fmt.Errorf("invalid point: %v", p)
		}
		if math.Signbit(p) {
			hasNegative = true
		}
		sum += p
	}
	t.sum, t.hasNegative = sum, hasNegative
	if t.frozen {
		t.points = append(slices.Clone(t.points), point...)
		t.frozen = false
	} else {
		t.points = append(t.points, point...)
	}
	return nil
}

// Count returns the number of points currently stored in the Sampdo instance.
func (t *Sampdo) Count() int {
	if t == nil {
		return 0
	}
	return len(t.points)
}

// CopyTo copies all points from the current Sampdo instance to the provided destination slice.
func (t *Sampdo) CopyTo(dst *[]float64) error {
	if t == nil {
		return fmt.Errorf("source is nil")
	}
	if dst == nil || *dst == nil {
		return fmt.Errorf("destination slice is nil")
	}
	*dst = append(*dst, t.points...)
	return nil
}

// AppendTo appends all points from the current Sampdo instance to the destination Sampdo instance.
func (t *Sampdo) AppendTo(dst *Sampdo) error {
	if t == nil {
		return fmt.Errorf("source is nil")
	}
	if dst == nil {
		return fmt.Errorf("destination is nil")
	}
	return dst.Append(t.points...)
}

func (t *Sampdo) Sorted() (*Sorted, error) {
	if t == nil {
		return nil, fmt.Errorf("sorted is nil")
	}
	if len(t.points) == 0 {
		return nil, fmt.Errorf("no points to sort")
	}
	t.frozen = true
	if t.hasNegative || t.PreferSlicesSort {
		slices.Sort(t.points)
	} else {
		radixSort(t.points)
	}
	return &Sorted{
		points:   t.points,
		sum:      t.sum,
		sumValid: true,
	}, nil
}
