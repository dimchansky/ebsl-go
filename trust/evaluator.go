package trust

import (
	"errors"

	"github.com/dimchansky/ebsl-go/opinion"
)

var (
	ErrInvalidExpression = errors.New("trust: invalid expression")
)

type ExpressionContext interface {
	R(link Link) *opinion.Type
	A(link Link) *opinion.Type
	G(*opinion.Type) float64
}

func EvaluateExpression(context ExpressionContext, expression Expression) (*opinion.Type, error) {
	ev := &expressionEvaluator{context: context}
	expression.Accept(ev)
	if ev.err != nil {
		return nil, ev.err
	}
	if ev.state != evaluated {
		return nil, ErrInvalidExpression
	}
	return ev.result, nil
}

type evaluatorState int

const (
	notEvaluated evaluatorState = iota
	evaluated    evaluatorState = iota
	consensus    evaluatorState = iota
)

type expressionEvaluator struct {
	context ExpressionContext
	result  *opinion.Type
	err     error
	state   evaluatorState
}

func (ev *expressionEvaluator) VisitFullUncertainty() {
	if ev.err != nil {
		return
	}

	switch ev.state {
	case notEvaluated:
		ev.result = opinion.FullUncertainty()
		ev.state = evaluated
	default:
		ev.err = ErrInvalidExpression
		return
	}
}

func (ev *expressionEvaluator) VisitDiscountingRule(r Link, a Link) {
	if ev.err != nil {
		return
	}

	switch ev.state {
	case notEvaluated:
		ctx := ev.context

		alpha := ctx.G(ctx.R(r))
		aOp := *ctx.A(a)

		ev.result = aOp.Mul(alpha)
		ev.state = evaluated
	case consensus:
		ctx := ev.context

		alpha := ctx.G(ctx.R(r))

		ev.result.PlusMul(alpha, ctx.A(a))
	default:
		ev.err = ErrInvalidExpression
		return
	}
}

func (ev *expressionEvaluator) VisitDirectReferralTrust(a Link) {
	if ev.err != nil {
		return
	}

	switch ev.state {
	case notEvaluated:
		ctx := ev.context

		aOp := *ctx.A(a)

		ev.result = &aOp
		ev.state = evaluated
	case consensus:
		ctx := ev.context

		ev.result.Plus(ctx.A(a))
	default:
		ev.err = ErrInvalidExpression
		return
	}
}

func (ev *expressionEvaluator) VisitConsensusListStart(count int) {
	if ev.err != nil {
		return
	}
	if ev.state != notEvaluated {
		ev.err = ErrInvalidExpression
		return
	}
	ev.state = consensus
	ev.result = opinion.FullUncertainty()
}

func (ev *expressionEvaluator) VisitConsensusList(index int, equation Expression) {
	if ev.err != nil {
		return
	}
	if ev.state != consensus {
		ev.err = ErrInvalidExpression
		return
	}

	equation.Accept(ev)
}

func (ev *expressionEvaluator) VisitConsensusListEnd() {
	if ev.err != nil {
		return
	}
	if ev.state != consensus {
		ev.err = ErrInvalidExpression
		return
	}
	ev.state = evaluated
}
