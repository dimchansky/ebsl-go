package equations

import (
	"sort"

	"github.com/dimchansky/ebsl-go/opinion"
	"github.com/dimchansky/ebsl-go/trust"
)

// ExpressionVisitor is a visitor for Expression
type ExpressionVisitor interface {
	VisitFullUncertainty() error
	VisitDiscountingRule(r trust.Link, a trust.Link) error
	VisitDirectReferralTrust(a trust.Link) error
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

// Equation represents final referral trust equation: R[i, j] = Expression
type Equation struct {
	// R represents R[i, j] - i's (possibly indirect) opinion about the trustworthiness of j
	R trust.Link
	// Expression of R[i, j]
	Expression Expression
}

// Equations is a set of final referral trust equation
type Equations []*Equation

// ExpressionContext to evaluate expression of final referral trust equation
type ExpressionContext interface {
	GetDirectReferralTrust(link trust.Link) *opinion.Type
	GetFinalReferralTrust(link trust.Link) *opinion.Type
	GetDiscount(*opinion.Type) float64
}

// Context to evaluate final referral trust equation
type Context interface {
	ExpressionContext
	// SetFinalReferralTrust used to update evaluated expression value
	SetFinalReferralTrust(link trust.Link, value *opinion.Type)
}

// Evaluate evaluates new final referral value from equation expression and updates final referral trust with the new value.
func (e *Equation) Evaluate(context Context) (res *opinion.Type, err error) {
	res, err = EvaluateExpression(context, e.Expression)
	if err == nil {
		context.SetFinalReferralTrust(e.R, res)
	}
	return res, err
}

// CreateEquations creates equations for the final referral trust
func CreateEquations(links trust.IterableLinks) Equations {
	uniques, referralsTo := getUniquesAndReferralsTo(links)

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

	var rEquations Equations
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
					rExp = rExp.circlePlus(a{From: k, To: to})
				} else if k != to && // diagonal in A equal to full uncertainty
					isReachable(from, k) { // should exists path from "from" to "k"
					rExp = rExp.circlePlus(discountingRule{r{From: from, To: k}, a{From: k, To: to}})
				}
			}

			if !rExp.IsFullUncertainty() {
				rEquations = append(rEquations,
					&Equation{
						R:          trust.Link{From: from, To: to},
						Expression: rExp,
					})
			}
		}
	}

	// order equations by direct first, then by indices
	orderEquationsByDirectRefAndIndices(rEquations)

	return rEquations
}

// orderEquationsByDirectRefAndIndices orders equations so that direct referral equations go first and the all equations are ordered by indices of R
func orderEquationsByDirectRefAndIndices(rEquations Equations) {
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
}

type uint64Set map[uint64]bool

// getUniquesAndReferralsTo collects all uniques from the links and creates index with referrals to each node
func getUniquesAndReferralsTo(links trust.IterableLinks) (uniques uint64Set, referralsTo map[uint64]uint64Set) {
	uniques = make(uint64Set)
	referralsTo = make(map[uint64]uint64Set)

	foreachLink := links.GetLinkIterator()
	// build graph with referrals and uniques
	_ = foreachLink(func(ref trust.Link) error {
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

		return nil
	})

	return
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
type r trust.Link

// direct referral trust A[i,j]
type a trust.Link

func (a a) circlePlus(p expression) expression {
	if p.IsFullUncertainty() {
		return a
	}
	res := []expression{a, p}
	return (*consensusList)(&res)
}

func (a a) IsFullUncertainty() bool          { return false }
func (a a) IsDiscountingRule() bool          { return false }
func (a a) IsDirectReferralTrust() bool      { return true }
func (a a) IsConsensusList() bool            { return false }
func (a a) Accept(v ExpressionVisitor) error { return v.VisitDirectReferralTrust(trust.Link(a)) }

// discountingRule R[i,j] ⊠ A[i,j]
type discountingRule struct {
	r r
	a a
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
	return v.VisitDiscountingRule(trust.Link(d.r), trust.Link(d.a))
}
