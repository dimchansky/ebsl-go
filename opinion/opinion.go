package opinion

import (
	"math"

	"github.com/dimchansky/ebsl-go/evidence"
)

type Type struct {
	B, D, U float64
}

func New(b, d, u float64) *Type {
	return &Type{B: b, D: d, U: u}
}

func FromEvidence(c uint64, e *evidence.Type) *Type {
	k := float64(c + e.P + e.N)
	return &Type{
		B: float64(e.P) / k,
		D: float64(e.N) / k,
		U: float64(c) / k,
	}
}

func (x *Type) ToEvidence(c uint64) *evidence.Type {
	p := math.Round(float64(c) * x.B / x.U)
	n := math.Round(float64(c) * x.D / x.U)
	return evidence.New(uint64(p), uint64(n))
}

func FullBelief() *Type {
	return &Type{1, 0, 0}
}

func FullDisbelief() *Type {
	return &Type{0, 1, 0}
}

func FullUncertainty() *Type {
	return &Type{0, 0, 1}
}

// Mul sets x to the scalar multiplication α·x and returns x.
func (x *Type) Mul(α float64) *Type {
	b := α * x.B
	d := α * x.D
	u := x.U
	k := b + d + u
	x.B = b / k
	x.D = d / k
	x.U = u / k
	return x
}

// Plus sets x to the x⊕y and returns x.
func (x *Type) Plus(y *Type) *Type {
	xu := x.U
	yu := y.U
	k := xu + yu - xu*yu
	x.B = (xu*y.B + yu*x.B) / k
	x.D = (xu*y.D + yu*x.D) / k
	x.U = xu * yu / k
	return x
}

// x⊠y
