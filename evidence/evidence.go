package evidence

type Type struct {
	P, N uint64
}

func New(p, n uint64) *Type {
	return &Type{P: p, N: n}
}