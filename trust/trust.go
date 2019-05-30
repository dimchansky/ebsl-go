package trust

import (
	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/opinion"
)

type Link struct {
	From uint64
	To   uint64
}

// EquationVisitor is A visitor for Equation
type EquationVisitor interface {
	VisitFullUncertainty()
	VisitDiscountingRule(r Link, a Link)
	VisitDirectReferralTrust(a Link)
	VisitConsensusList(index int, equation Equation)
}

// Equation represents final referral trust equation
type Equation interface {
	IsFullUncertainty() bool
	IsDiscountingRule() bool
	IsDirectReferralTrust() bool
	IsConsensusList() bool
	Accept(v EquationVisitor)
}

// FinalReferralTrustEquation represents final reference trust equation: R = Equation
type FinalReferralTrustEquation struct {
	R        Link
	Equation Equation
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
					rEq = rEq.circlePlus(A{k, to})
				} else if k != to && // diagonal in A equal to full uncertainty
					isReachable(from, k) { // should exists path from "from" to "k"
					rEq = rEq.circlePlus(discountingRule{R{from, k}, A{k, to}})
				}
			}

			if !rEq.IsFullUncertainty() {
				rEquations = append(rEquations,
					&FinalReferralTrustEquation{
						R:        Link{from, to},
						Equation: rEq,
					})
			}
		}
	}

	return rEquations
}

// Equation represents final referral trust equation
type equation interface {
	Equation
	circlePlus(p equation) equation
}

// full uncertainty
type u struct{}

func (u) circlePlus(p equation) equation { return p }
func (u) IsFullUncertainty() bool        { return true }
func (u) IsDiscountingRule() bool        { return false }
func (u) IsDirectReferralTrust() bool    { return false }
func (u) IsConsensusList() bool          { return false }
func (u) Accept(v EquationVisitor)       { v.VisitFullUncertainty() }

// ⊕ (circlePlus operation on A list)
type consensusList []equation

func (l *consensusList) circlePlus(p equation) equation {
	if !p.IsFullUncertainty() {
		*l = append(*l, p)
	}
	return l
}

func (l *consensusList) IsFullUncertainty() bool     { return len(*l) == 0 }
func (l *consensusList) IsDiscountingRule() bool     { return false }
func (l *consensusList) IsDirectReferralTrust() bool { return false }
func (l *consensusList) IsConsensusList() bool       { return true }
func (l *consensusList) Accept(v EquationVisitor) {
	for idx, value := range *l {
		v.VisitConsensusList(idx, value)
	}
}

// final referral trust R[i,j]
type R Link

// direct referral trust A[i,j]
type A Link

func (a A) circlePlus(p equation) equation {
	if p.IsFullUncertainty() {
		return a
	}
	res := []equation{a, p}
	return (*consensusList)(&res)
}

func (a A) IsFullUncertainty() bool     { return false }
func (a A) IsDiscountingRule() bool     { return false }
func (a A) IsDirectReferralTrust() bool { return true }
func (a A) IsConsensusList() bool       { return false }
func (a A) Accept(v EquationVisitor)    { v.VisitDirectReferralTrust(Link(a)) }

// discountingRule R[i,j] ⊠ A[i,j]
type discountingRule struct {
	r R
	a A
}

func (d discountingRule) circlePlus(p equation) equation {
	res := []equation{d, p}
	return (*consensusList)(&res)
}

func (d discountingRule) IsFullUncertainty() bool     { return false }
func (d discountingRule) IsDiscountingRule() bool     { return true }
func (d discountingRule) IsDirectReferralTrust() bool { return false }
func (d discountingRule) IsConsensusList() bool       { return false }
func (d discountingRule) Accept(v EquationVisitor)    { v.VisitDiscountingRule(Link(d.r), Link(d.a)) }
