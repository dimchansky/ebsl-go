package opinion

import (
	"fmt"

	"github.com/dimchansky/ebsl-go/evidence"
)

// Type of opinion
type Type struct {
	B, D, U float64
}

// String implements fmt.Stringer
func (x *Type) String() string {
	return fmt.Sprintf("{B: %v, D: %v, U: %v}", x.B, x.D, x.U)
}

// New creates new instance of opinion
func New(b, d, u float64) *Type {
	return &Type{B: b, D: d, U: u}
}

// FromEvidence converts evidence to opinion using `c` as soft threshold/"unit" of evidence (must be positive number).
func FromEvidence(c uint64, e *evidence.Type) *Type { return new(Type).FromEvidence(c, e) }

// FromEvidence converts evidence and updates opinion using `c` as soft threshold/"unit" of evidence (must be positive number).
// Method returns updated value.
func (x *Type) FromEvidence(c uint64, e *evidence.Type) *Type {
	k := float64(c) + e.P + e.N
	x.B = e.P / k
	x.D = e.N / k
	x.U = float64(c) / k
	return x
}

// ToEvidence converts opinion to evidence using `c` as soft threshold/"unit" of evidence (must be positive number).
func (x *Type) ToEvidence(c uint64) *evidence.Type {
	p := float64(c) * x.B / x.U
	n := float64(c) * x.D / x.U
	return evidence.New(p, n)
}

// FullBelief returns full belief opinion
func FullBelief() *Type {
	return &Type{1, 0, 0}
}

// FullDisbelief returns full disbelief opinion
func FullDisbelief() *Type {
	return &Type{0, 1, 0}
}

// FullUncertainty returns full uncertainty opinion
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

// PlusMul sets x to the x⊕(α·y) and returns x.
func (x *Type) PlusMul(α float64, y *Type) *Type {
	if α == 0 {
		return x
	}
	xu := x.U
	yu := y.U
	k := yu + α*xu*(1-yu)
	x.B = (α*xu*y.B + yu*x.B) / k
	x.D = (α*xu*y.D + yu*x.D) / k
	x.U = xu * yu / k
	return x
}
