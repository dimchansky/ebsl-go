package equations

import (
	"fmt"

	"github.com/dimchansky/ebsl-go/opinion"
	"github.com/dimchansky/ebsl-go/trust"
)

// FinalReferralTrustExpressionVisitor is a visitor for FinalReferralTrustExpression
type FinalReferralTrustExpressionVisitor interface {
	VisitFullUncertainty() error
	VisitDiscountingRule(r trust.Link, a trust.Link) error
	VisitDirectReferralTrust(a trust.Link) error
	VisitConsensusListStart(count int) error
	VisitConsensusList(index int, equation FinalReferralTrustExpression) error
	VisitConsensusListEnd() error
}

// FinalReferralTrustExpression represents final referral trust expression
type FinalReferralTrustExpression interface {
	IsFullUncertainty() bool
	IsDiscountingRule() bool
	IsDirectReferralTrust() bool
	IsConsensusList() bool
	Accept(v FinalReferralTrustExpressionVisitor) error
}

// FinalReferralTrustEquation represents final referral trust equation: R[i, j] = FinalReferralTrustExpression
type FinalReferralTrustEquation struct {
	// R represents R[i, j] - i's (possibly indirect) opinion about the trustworthiness of j
	R trust.Link
	// FinalReferralTrustExpression of R[i, j]
	Expression FinalReferralTrustExpression
}

// FinalReferralTrustEquations is a set of final referral trust equation
type FinalReferralTrustEquations []*FinalReferralTrustEquation

// FinalReferralTrustExpressionContext to evaluate expression of final referral trust equation
type FinalReferralTrustExpressionContext interface {
	GetDirectReferralTrust(link trust.Link) *opinion.Type
	GetFinalReferralTrust(link trust.Link) *opinion.Type
	GetDiscount(*opinion.Type) float64
}

// FinalReferralTrustEquationContext to evaluate final referral trust equation
type FinalReferralTrustEquationContext interface {
	FinalReferralTrustExpressionContext
	// SetFinalReferralTrust used to update evaluated expression value
	SetFinalReferralTrust(link trust.Link, value *opinion.Type)
}

// EvaluateFinalReferralTrust evaluates new final referral value from equation expression and updates final referral trust with the new value.
func (e *FinalReferralTrustEquation) EvaluateFinalReferralTrust(context FinalReferralTrustEquationContext) (res *opinion.Type, err error) {
	res, err = EvaluateFinalReferralTrustExpression(context, e.Expression)
	if err == nil {
		context.SetFinalReferralTrust(e.R, res)
	}
	return res, err
}

type DefaultContext struct {
	DirectReferralTrust trust.DirectReferralOpinion
	FinalReferralTrust  trust.FinalReferralOpinion
}

func NewDefaultContext(a trust.DirectReferralOpinion) *DefaultContext {
	return &DefaultContext{
		DirectReferralTrust: a,
		FinalReferralTrust:  make(trust.FinalReferralOpinion),
	}
}

func (c *DefaultContext) GetDirectReferralTrust(link trust.Link) *opinion.Type {
	res, ok := c.DirectReferralTrust[link]
	if !ok {
		panic(fmt.Sprintf("direct referral trust not found: [%v, %v]", link.From, link.To))
	}
	return res
}

func (c *DefaultContext) GetFinalReferralTrust(link trust.Link) *opinion.Type {
	if res, ok := c.FinalReferralTrust[link]; ok {
		return res
	}
	return opinion.FullBelief()
}

func (c *DefaultContext) GetDiscount(o *opinion.Type) float64 {
	return o.B
}

func (c *DefaultContext) SetFinalReferralTrust(link trust.Link, value *opinion.Type) {
	c.FinalReferralTrust[link] = value
}

// CreateEquations creates equations for the final referral trust
func CreateEquations(links trust.IterableLinks) FinalReferralTrustEquations {
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

	var rEquations FinalReferralTrustEquations
	// generate equations for final referral trust (R)
	for from := range uniques {
		for to, referrals := range referralsTo {
			// construct FinalReferralTrustExpression for R[from,to]
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
					&FinalReferralTrustEquation{
						R:          trust.Link{From: from, To: to},
						Expression: rExp,
					})
			}
		}
	}

	return rEquations
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

// FinalReferralTrustExpression represents final referral trust expression
type expression interface {
	FinalReferralTrustExpression
	circlePlus(p expression) expression
}

// full uncertainty
type u struct{}

func (u) circlePlus(p expression) expression                 { return p }
func (u) IsFullUncertainty() bool                            { return true }
func (u) IsDiscountingRule() bool                            { return false }
func (u) IsDirectReferralTrust() bool                        { return false }
func (u) IsConsensusList() bool                              { return false }
func (u) Accept(v FinalReferralTrustExpressionVisitor) error { return v.VisitFullUncertainty() }

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
func (l *consensusList) Accept(v FinalReferralTrustExpressionVisitor) (err error) {
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

func (a a) IsFullUncertainty() bool     { return false }
func (a a) IsDiscountingRule() bool     { return false }
func (a a) IsDirectReferralTrust() bool { return true }
func (a a) IsConsensusList() bool       { return false }
func (a a) Accept(v FinalReferralTrustExpressionVisitor) error {
	return v.VisitDirectReferralTrust(trust.Link(a))
}

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
func (d discountingRule) Accept(v FinalReferralTrustExpressionVisitor) error {
	return v.VisitDiscountingRule(trust.Link(d.r), trust.Link(d.a))
}
