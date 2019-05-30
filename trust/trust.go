package trust

import (
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

func (dro DirectReferralOpinion) CreateFinalReferralTrustEquations() []*FinalReferralTrustEquation {
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

	var rEquations []*FinalReferralTrustEquation
	// generate equations for final referral trust (R)
	for from := range uniques {
		for to, referrals := range referralsTo {
			// construct Equation for R[from,to]
			if from == to {
				// R[from,from] = full belief (skip it)
				continue
			}
			var rEq equation = u{}
			for k := range referrals {
				if k == from { // diagonal in R equal to full belief
					rEq = rEq.plus(A{k, to})
				} else if k != to && // diagonal in A equal to full uncertainty
					isReachable(from, k) { // should exists path from "from" to "k"
					rEq = rEq.plus(discountingRule{R{from, k}, A{k, to}})
				}
			}

			if !rEq.isFullUncertainty() {
				rEquations = append(rEquations,
					&FinalReferralTrustEquation{
						R:        R{from, to},
						Equation: rEq,
					})
			}
		}
	}

	return rEquations
}

// EquationVisitor is A visitor for Equation
type EquationVisitor interface {
	VisitFullUncertainty()
	VisitDiscountingRule(r R, a A)
	VisitDirectReferralTrust(a A)
	VisitConsensusStart()
	VisitConsensusOpinion(index int, equation Equation)
	VisitConsensusEnd()
}

// Equation represents final referral trust equation
type Equation interface {
	Accept(v EquationVisitor)
}

// Equation represents final referral trust equation
type equation interface {
	Equation
	plus(p equation) equation
	isFullUncertainty() bool
}

// FinalReferralTrustEquation represents final reference trust equation: R = Equation
type FinalReferralTrustEquation struct {
	R        R
	Equation Equation
}

// full uncertainty
type u struct{}

func (u) plus(p equation) equation { return p }
func (u) isFullUncertainty() bool  { return true }
func (u) Accept(v EquationVisitor) { v.VisitFullUncertainty() }

// ⊕ (plus operation on A list)
type consensusList []equation

func (l *consensusList) plus(p equation) equation {
	if !p.isFullUncertainty() {
		*l = append(*l, p)
	}
	return l
}

func (l *consensusList) isFullUncertainty() bool { return len(*l) == 0 }

func (l *consensusList) Accept(v EquationVisitor) {
	v.VisitConsensusStart()
	for idx, value := range *l {
		v.VisitConsensusOpinion(idx, value)
	}
	v.VisitConsensusEnd()
}

// final referral trust R[i,j]
type R Link

// direct referral trust A[i,j]
type A Link

func (a A) plus(p equation) equation {
	if p.isFullUncertainty() {
		return a
	}
	res := []equation{a, p}
	return (*consensusList)(&res)
}

func (a A) isFullUncertainty() bool { return false }

func (a A) Accept(v EquationVisitor) { v.VisitDirectReferralTrust(a) }

// discountingRule R[i,j] ⊠ A[i,j]
type discountingRule struct {
	r R
	a A
}

func (d discountingRule) plus(p equation) equation {
	res := []equation{d, p}
	return (*consensusList)(&res)
}

func (d discountingRule) isFullUncertainty() bool { return false }

func (d discountingRule) Accept(v EquationVisitor) { v.VisitDiscountingRule(d.r, d.a) }
