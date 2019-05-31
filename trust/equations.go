package trust

import "sort"

// ExpressionVisitor is A visitor for Expression
type ExpressionVisitor interface {
	VisitFullUncertainty() error
	VisitDiscountingRule(r Link, a Link) error
	VisitDirectReferralTrust(a Link) error
	VisitConsensusListStart(count int) error
	VisitConsensusList(index int, equation Expression) error
	VisitConsensusListEnd() error
}

// Expression represents final referral trust expression
type Expression interface {
	IsFullUncertainty() bool
	IsDiscountingRule() bool
	IsDirectReferralTrust() bool
	IsConsensusList() bool
	Accept(v ExpressionVisitor) error
}

// FinalReferralTrustEquation represents final reference trust expression: R = Expression
type FinalReferralTrustEquation struct {
	R          Link
	Expression Expression
}

func CreateFinalReferralTrustEquations(dro DirectReferralOpinion) []*FinalReferralTrustEquation {
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
			// construct Expression for R[from,to]
			if from == to {
				// R[from,from] = full belief (skip it)
				continue
			}
			var rExp expression = u{}
			for k := range referrals {
				if k == from { // diagonal in R equal to full belief
					rExp = rExp.circlePlus(A{k, to})
				} else if k != to && // diagonal in A equal to full uncertainty
					isReachable(from, k) { // should exists path from "from" to "k"
					rExp = rExp.circlePlus(discountingRule{R{from, k}, A{k, to}})
				}
			}

			if !rExp.IsFullUncertainty() {
				rEquations = append(rEquations,
					&FinalReferralTrustEquation{
						R:          Link{from, to},
						Expression: rExp,
					})
			}
		}
	}

	// order equations by direct first, then by indices
	sort.Slice(rEquations, func(i, j int) bool {
		iEq := rEquations[i]
		jEq := rEquations[j]
		iExpDirect := iEq.Expression.IsDirectReferralTrust()
		jExpDirect := jEq.Expression.IsDirectReferralTrust()
		if iExpDirect != jExpDirect {
			return iExpDirect // direct equations go first
		}

		// the sort by R indices
		iFrom := iEq.R.From
		jFrom := jEq.R.From
		if iFrom != jFrom {
			return iFrom < jFrom
		}

		return iEq.R.To < jEq.R.To
	})

	return rEquations
}

// Expression represents final referral trust expression
type expression interface {
	Expression
	circlePlus(p expression) expression
}

// full uncertainty
type u struct{}

func (u) circlePlus(p expression) expression { return p }
func (u) IsFullUncertainty() bool            { return true }
func (u) IsDiscountingRule() bool            { return false }
func (u) IsDirectReferralTrust() bool        { return false }
func (u) IsConsensusList() bool              { return false }
func (u) Accept(v ExpressionVisitor) error   { return v.VisitFullUncertainty() }

// ⊕ (circlePlus operation on A list)
type consensusList []expression

func (l *consensusList) circlePlus(p expression) expression {
	if !p.IsFullUncertainty() {
		*l = append(*l, p)
	}
	return l
}

func (l *consensusList) IsFullUncertainty() bool     { return len(*l) == 0 }
func (l *consensusList) IsDiscountingRule() bool     { return false }
func (l *consensusList) IsDirectReferralTrust() bool { return false }
func (l *consensusList) IsConsensusList() bool       { return true }
func (l *consensusList) Accept(v ExpressionVisitor) (err error) {
	if err = v.VisitConsensusListStart(len(*l)); err != nil {
		return
	}
	for idx, value := range *l {
		if err = v.VisitConsensusList(idx, value); err != nil {
			return
		}
	}
	return v.VisitConsensusListEnd()
}

// final referral trust R[i,j]
type R Link

// direct referral trust A[i,j]
type A Link

func (a A) circlePlus(p expression) expression {
	if p.IsFullUncertainty() {
		return a
	}
	res := []expression{a, p}
	return (*consensusList)(&res)
}

func (a A) IsFullUncertainty() bool          { return false }
func (a A) IsDiscountingRule() bool          { return false }
func (a A) IsDirectReferralTrust() bool      { return true }
func (a A) IsConsensusList() bool            { return false }
func (a A) Accept(v ExpressionVisitor) error { return v.VisitDirectReferralTrust(Link(a)) }

// discountingRule R[i,j] ⊠ A[i,j]
type discountingRule struct {
	r R
	a A
}

func (d discountingRule) circlePlus(p expression) expression {
	res := []expression{d, p}
	return (*consensusList)(&res)
}

func (d discountingRule) IsFullUncertainty() bool     { return false }
func (d discountingRule) IsDiscountingRule() bool     { return true }
func (d discountingRule) IsDirectReferralTrust() bool { return false }
func (d discountingRule) IsConsensusList() bool       { return false }
func (d discountingRule) Accept(v ExpressionVisitor) error {
	return v.VisitDiscountingRule(Link(d.r), Link(d.a))
}
