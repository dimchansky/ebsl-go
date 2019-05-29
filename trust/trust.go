package trust

import (
	"fmt"
	"strings"

	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/opinion"
)

type Link struct {
	From uint64
	To   uint64
}

type DirectReferralEvidence map[Link]*evidence.Type

func (dre DirectReferralEvidence) ToDirectReferralOpinion(c uint64) DirectReferralOpinion {
	res := make(DirectReferralOpinion, len(dre))
	for ref, ev := range dre {
		res[ref] = opinion.FromEvidence(c, ev)
	}
	return res
}

type DirectReferralOpinion map[Link]*opinion.Type

func (dro DirectReferralOpinion) XXX() {
	type uint64Set map[uint64]bool

	uniques := make(uint64Set)
	referralsTo := make(map[uint64]uint64Set)

	for ref := range dro {
		from := ref.From
		to := ref.To

		if !uniques[to] {
			uniques[to] = true
		}
		if !uniques[from] {
			uniques[from] = true
		}

		referrals := referralsTo[to]
		if referrals == nil {
			referrals = make(uint64Set)
			referralsTo[to] = referrals
		}
		referrals[from] = true
	}

	// TODO: optimize for all nodes
	isReachable := func(from, to uint64) bool {
		if from == to {
			return true
		}

		for visited, stack := map[uint64]bool{to: true}, []uint64{to}; len(stack) > 0; {
			n := len(stack) - 1
			to = stack[n]
			stack = stack[:n]

			for i := range referralsTo[to] {
				if i == from {
					return true
				}

				if !visited[i] {
					visited[i] = true
					stack = append(stack, i)
				}
			}
		}

		return false
	}

	rEquations := make(map[r]Equation)
	// generate equations for final referral trust (R)
	for from := range uniques {
		for to, referrals := range referralsTo {
			// construct Equation for R[from,to]
			if from == to {
				// R[from,from] = full belief (skip it)
				continue
			}
			var rEq Equation = u{}
			for k := range referrals {
				if k == from { // diagonal in R equal to full belief
					rEq = rEq.Plus(a{k, to})
				} else if k != to && // diagonal in A equal to full uncertainty
					isReachable(from, k) { // should exists path from "from" to "k"
					rEq = rEq.Plus(discount{r{from, k}, a{k, to}})
				}
			}

			if !rEq.IsFullUncertainty() {
				fmt.Println(r{from, to}, "=", rEq)
				rEquations[r{from, to}] = rEq
			}
		}
	}
}

// Equation represents final referral trust equation
type Equation interface {
	fmt.Stringer
	Plus(p Equation) Equation
	IsFullUncertainty() bool
}

// full uncertainty
type u struct{}

func (u) String() string           { return "U" }
func (u) Plus(p Equation) Equation { return p }
func (u) IsFullUncertainty() bool  { return true }

// ⊗ (plus operation on a list)
type plusList []Equation

func (l *plusList) String() string {
	ss := make([]string, len(*l))
	for i, p := range *l {
		ss[i] = p.String()
	}
	return strings.Join(ss, " ⊗ ")
}

func (l *plusList) Plus(p Equation) Equation {
	if !p.IsFullUncertainty() {
		*l = append(*l, p)
	}
	return l
}

func (l *plusList) IsFullUncertainty() bool { return len(*l) == 0 }

// final referral trust R[i,j]
type r Link

func (r r) String() string { return fmt.Sprintf("R[%v,%v]", r.From, r.To) }

// direct referral trust A[i,j]
type a Link

func (a a) String() string { return fmt.Sprintf("A[%v,%v]", a.From, a.To) }

func (a a) Plus(p Equation) Equation {
	if p.IsFullUncertainty() {
		return a
	}
	res := []Equation{a, p}
	return (*plusList)(&res)
}

func (a a) IsFullUncertainty() bool { return false }

// discount R[i,j] ⊗ A[i,j]
type discount struct {
	r r
	a a
}

func (d discount) String() string {
	return fmt.Sprintf("(%v ⊗ %v)", d.r.String(), d.a.String())
}

func (d discount) Plus(p Equation) Equation {
	res := []Equation{d, p}
	return (*plusList)(&res)
}

func (d discount) IsFullUncertainty() bool { return false }
