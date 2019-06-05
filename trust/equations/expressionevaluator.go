package equations

import (
	"errors"

	"github.com/dimchansky/ebsl-go/opinion"
	"github.com/dimchansky/ebsl-go/trust"
)

var (
	// ErrInvalidExpression is returned when expression is invalid and cannot be evaluated
	ErrInvalidExpression = errors.New("trust: invalid expression")
)

// EvaluateFinalReferralTrustExpression evaluates expression using expression context and returns evaluated value or error
func EvaluateFinalReferralTrustExpression(context FinalReferralTrustExpressionContext, expression FinalReferralTrustExpression) (*opinion.Type, error) {
	ev := &frtExpressionEvaluator{context: context}
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

type frtExpressionEvaluator struct {
	context FinalReferralTrustExpressionContext
	result  *opinion.Type
	state   evaluatorState
}

func (ev *frtExpressionEvaluator) VisitFullUncertainty() error {
	if ev.state != notEvaluated {
		return ErrInvalidExpression
	}

	ev.result = opinion.FullUncertainty()
	ev.state = evaluated
	return nil
}

func (ev *frtExpressionEvaluator) VisitDiscountingRule(r trust.Link, a trust.Link) (err error) {
	switch ev.state {
	case notEvaluated:
		ctx := ev.context

		alpha := ctx.GetDiscount(ctx.GetFinalReferralTrust(r))
		aOp := *ctx.GetDirectReferralTrust(a)

		ev.result = aOp.Mul(alpha)
		ev.state = evaluated
	case consensus:
		ctx := ev.context

		alpha := ctx.GetDiscount(ctx.GetFinalReferralTrust(r))

		ev.result.PlusMul(alpha, ctx.GetDirectReferralTrust(a))
	default:
		err = ErrInvalidExpression
	}
	return
}

func (ev *frtExpressionEvaluator) VisitDirectReferralTrust(a trust.Link) (err error) {
	switch ev.state {
	case notEvaluated:
		ctx := ev.context

		aOp := *ctx.GetDirectReferralTrust(a)

		ev.result = &aOp
		ev.state = evaluated
	case consensus:
		ctx := ev.context

		ev.result.Plus(ctx.GetDirectReferralTrust(a))
	default:
		err = ErrInvalidExpression
	}
	return
}

func (ev *frtExpressionEvaluator) VisitConsensusListStart(count int) error {
	if ev.state != notEvaluated {
		return ErrInvalidExpression
	}

	ev.state = consensus
	ev.result = opinion.FullUncertainty()
	return nil
}

func (ev *frtExpressionEvaluator) VisitConsensusList(index int, equation FinalReferralTrustExpression) error {
	if ev.state != consensus {
		return ErrInvalidExpression
	}

	return equation.Accept(ev)
}

func (ev *frtExpressionEvaluator) VisitConsensusListEnd() error {
	if ev.state != consensus {
		return ErrInvalidExpression
	}

	ev.state = evaluated
	return nil
}
