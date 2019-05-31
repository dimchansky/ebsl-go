package equations

import (
	"errors"

	"github.com/dimchansky/ebsl-go/opinion"
	"github.com/dimchansky/ebsl-go/trust"
)

var (
	ErrInvalidExpression = errors.New("trust: invalid expression")
)

type ExpressionContext interface {
	R(link trust.Link) *opinion.Type
	A(link trust.Link) *opinion.Type
	G(*opinion.Type) float64
}

func EvaluateExpression(context ExpressionContext, expression Expression) (*opinion.Type, error) {
	ev := &expressionEvaluator{context: context}
	if err := expression.Accept(ev); err != nil {
		return nil, err
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
	state   evaluatorState
}

func (ev *expressionEvaluator) VisitFullUncertainty() error {
	if ev.state != notEvaluated {
		return ErrInvalidExpression
	}

	ev.result = opinion.FullUncertainty()
	ev.state = evaluated
	return nil
}

func (ev *expressionEvaluator) VisitDiscountingRule(r trust.Link, a trust.Link) (err error) {
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
		err = ErrInvalidExpression
	}
	return
}

func (ev *expressionEvaluator) VisitDirectReferralTrust(a trust.Link) (err error) {
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
		err = ErrInvalidExpression
	}
	return
}

func (ev *expressionEvaluator) VisitConsensusListStart(count int) error {
	if ev.state != notEvaluated {
		return ErrInvalidExpression
	}

	ev.state = consensus
	ev.result = opinion.FullUncertainty()
	return nil
}

func (ev *expressionEvaluator) VisitConsensusList(index int, equation Expression) error {
	if ev.state != consensus {
		return ErrInvalidExpression
	}

	return equation.Accept(ev)
}

func (ev *expressionEvaluator) VisitConsensusListEnd() error {
	if ev.state != consensus {
		return ErrInvalidExpression
	}

	ev.state = evaluated
	return nil
}
