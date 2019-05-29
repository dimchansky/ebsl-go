package evidence

import "fmt"

type Type struct {
	P, N float64
}

func (e *Type) String() string {
	return fmt.Sprintf("{P: %v, N: %v}", e.P, e.N)
}

func New(p, n float64) *Type {
	return &Type{P: p, N: n}
}

