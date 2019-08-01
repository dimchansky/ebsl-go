package evidence

import "fmt"

// Type of evidence
type Type struct {
	P, N float64
}

// String implements fmt.Stringer
func (e Type) String() string {
	return fmt.Sprintf("{P: %v, N: %v}", e.P, e.N)
}

// New creates new instance of evidence
func New(p, n float64) Type {
	return Type{P: p, N: n}
}
