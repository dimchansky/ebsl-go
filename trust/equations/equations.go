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

// FinalFunctionalTrustContext is a context to evaluate final functional trust
type FinalFunctionalTrustContext interface {
	GetFinalReferralTrust(link trust.Link) *opinion.Type
	GetDiscount(*opinion.Type) float64
}

// FinalReferralTrustExpressionContext to evaluate expression of final referral trust equation
type FinalReferralTrustExpressionContext interface {
	FinalFunctionalTrustContext
	GetDirectReferralTrust(link trust.Link) *opinion.Type
}

// FinalReferralTrustEquationContext to evaluate final referral trust equation
type FinalReferralTrustEquationContext interface {
	FinalReferralTrustExpressionContext
	// SetFinalReferralTrust used to update evaluated expression value
	SetFinalReferralTrust(link trust.Link, value *opinion.Type)
}

// FinalReferralTrustEquations is a set of final referral trust equation
type FinalReferralTrustEquations []*FinalReferralTrustEquation

// EvaluateFinalReferralTrust evaluates new final referral value from equation expression and updates final referral trust with the new value.
func (e *FinalReferralTrustEquation) EvaluateFinalReferralTrust(context FinalReferralTrustEquationContext) (res *opinion.Type, err error) {
	res, err = EvaluateFinalReferralTrustExpression(context, e.Expression)
	if err == nil {
		context.SetFinalReferralTrust(e.R, res)
	}
	return res, err
}

func EvaluateFinalFunctionalTrust(ctx FinalFunctionalTrustContext, of uint64, dft trust.DirectFunctionalTrust) *opinion.Type {
	res := opinion.FullUncertainty()

	for opinionOf, directOpinion := range dft {
		alpha := ctx.GetDiscount(ctx.GetFinalReferralTrust(trust.Link{From: of, To: opinionOf}))

		res.PlusMul(alpha, directOpinion)
	}

	return res
}

type DefaultFinalReferralTrustEquationContext struct {
	DirectReferralTrust trust.DirectReferralOpinion
	FinalReferralTrust  trust.FinalReferralOpinion
}

func NewDefaultFinalReferralTrustEquationContext(a trust.DirectReferralOpinion) *DefaultFinalReferralTrustEquationContext {
	return &DefaultFinalReferralTrustEquationContext{
		DirectReferralTrust: a,
		FinalReferralTrust:  make(trust.FinalReferralOpinion),
	}
}

func (c *DefaultFinalReferralTrustEquationContext) GetDirectReferralTrust(link trust.Link) *opinion.Type {
	res, ok := c.DirectReferralTrust[link]
	if !ok {
		panic(fmt.Sprintf("direct referral trust not found: [%v, %v]", link.From, link.To))
	}
	return res
}

func (c *DefaultFinalReferralTrustEquationContext) GetFinalReferralTrust(link trust.Link) *opinion.Type {
	if res, ok := c.FinalReferralTrust[link]; ok {
		return res
	}
	return opinion.FullBelief()
}

func (c *DefaultFinalReferralTrustEquationContext) GetDiscount(o *opinion.Type) float64 {
	return o.B
}

func (c *DefaultFinalReferralTrustEquationContext) SetFinalReferralTrust(link trust.Link, value *opinion.Type) {
	c.FinalReferralTrust[link] = value
}

// CreateFinalReferralTrustEquations creates equations for the final referral trust
func CreateFinalReferralTrustEquations(links trust.IterableLinks) (rEquations FinalReferralTrustEquations) {
	sourceGraph, sinkGraph := buildGraph(links)

	stack := make([]uint64, 0, len(sinkGraph)) // reusable stack of nodes to visit
	for from := range sourceGraph {
		// mark all isReachable nodes from current node
		isReachable := map[uint64]bool{from: true}
		stack = append(stack, from)
		for len(stack) > 0 {
			n := len(stack) - 1
			sourceNode := stack[n]
			stack = stack[:n]

			sinkNodes := sourceGraph[sourceNode]
			for sinkNode := range sinkNodes {
				if !isReachable[sinkNode] {
					isReachable[sinkNode] = true
					stack = append(stack, sinkNode)
				}
			}
		}

		// generate equations for final referral trust (R)
		for to := range isReachable {
			if from == to {
				// R[from,from] = full belief (skip it)
				continue
			}

			var rExp expression = u{}
			for k := range sinkGraph[to] {
				if k == from { // diagonal in R equal to full belief
					rExp = rExp.circlePlus(a{From: k, To: to})
				} else if k != to && // diagonal in A equal to full uncertainty
					isReachable[k] { // should exists path from "from" to "k"
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

	return
}

type uint64Set map[uint64]bool

// in source graph all keys are source vertexes and values are sink vertexes
// in sink graph all keys are sink vertexes and values are source vertexes
func buildGraph(links trust.IterableLinks) (sourceGraph map[uint64]uint64Set, sinkGraph map[uint64]uint64Set) {
	sourceGraph = make(map[uint64]uint64Set)
	sinkGraph = make(map[uint64]uint64Set)

	foreachLink := links.GetLinkIterator()
	_ = foreachLink(func(ref trust.Link) error {
		from := ref.From
		to := ref.To

		sinkNodes := sourceGraph[from]
		if sinkNodes == nil {
			sinkNodes = make(uint64Set)
			sourceGraph[from] = sinkNodes
		}
		sinkNodes[to] = true

		sourceNodes := sinkGraph[to]
		if sourceNodes == nil {
			sourceNodes = make(uint64Set)
			sinkGraph[to] = sourceNodes
		}
		sourceNodes[from] = true

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
