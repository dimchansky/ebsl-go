package trust_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/trust"
	"github.com/dimchansky/ebsl-go/trust/equations"
	"github.com/dimchansky/ebsl-go/trust/equations/solver"
)

func TestExample(t *testing.T) {
	c := uint64(2)

	a := trust.DirectReferralEvidence{
		trust.Link{From: 1, To: 2}: evidence.New(2, 2),
		trust.Link{From: 2, To: 3}: evidence.New(2, 2),
		trust.Link{From: 3, To: 2}: evidence.New(2, 2),
	}.ToDirectReferralOpinion(c)

	eqs := equations.CreateEquations(a)
	logEquations(t, eqs)

	if err := solver.SolveEquations(
		equations.NewDefaultContext(a),
		eqs,
		solver.UseOnEpochEndCallback(func(epoch uint, aggregatedDistance float64) error {
			t.Logf("Epoch %v error: %v\n", epoch, aggregatedDistance)
			return nil
		}),
	); err != nil {
		t.Fatal(err)
	}
}

func TestExample2(t *testing.T) {
	c := uint64(2)

	a := trust.DirectReferralEvidence{
		trust.Link{From: 1, To: 2}: evidence.New(2, 2),
		trust.Link{From: 2, To: 3}: evidence.New(2, 2),
		trust.Link{From: 3, To: 4}: evidence.New(2, 2),
		trust.Link{From: 3, To: 5}: evidence.New(2, 2),
		trust.Link{From: 4, To: 5}: evidence.New(2, 2),
		trust.Link{From: 4, To: 6}: evidence.New(2, 2),
		trust.Link{From: 5, To: 6}: evidence.New(2, 2),
		trust.Link{From: 6, To: 7}: evidence.New(2, 2),
	}.ToDirectReferralOpinion(c)

	eqs := equations.CreateEquations(a)
	logEquations(t, eqs)

	if err := solver.SolveEquations(
		equations.NewDefaultContext(a),
		eqs,
		solver.UseOnEpochEndCallback(func(epoch uint, aggregatedDistance float64) error {
			t.Logf("Epoch %v error: %v\n", epoch, aggregatedDistance)
			return nil
		}),
	); err != nil {
		t.Fatal(err)
	}
}

func TestExample3(t *testing.T) {
	c := uint64(2)

	a := trust.DirectReferralEvidence{
		trust.Link{From: 1, To: 2}: evidence.New(2, 2),
		trust.Link{From: 2, To: 3}: evidence.New(2, 2),
		trust.Link{From: 3, To: 4}: evidence.New(2, 2),
		trust.Link{From: 4, To: 5}: evidence.New(2, 2),
		trust.Link{From: 5, To: 6}: evidence.New(2, 2),
		trust.Link{From: 6, To: 7}: evidence.New(2, 2),
		trust.Link{From: 7, To: 8}: evidence.New(2, 2),
		trust.Link{From: 8, To: 9}: evidence.New(2, 2),
	}.ToDirectReferralOpinion(c)

	eqs := equations.CreateEquations(a)
	logEquations(t, eqs)

	if err := solver.SolveEquations(
		equations.NewDefaultContext(a),
		eqs,
		solver.UseOnEpochEndCallback(func(epoch uint, aggregatedDistance float64) error {
			t.Logf("Epoch %v error: %v\n", epoch, aggregatedDistance)
			return nil
		}),
	); err != nil {
		t.Fatal(err)
	}
}

func logEquations(t *testing.T, eqs []*equations.Equation) {
	for _, eq := range eqs {
		t.Log(frtEqToString(eq))
	}
}

func rToString(r trust.Link) string { return fmt.Sprintf("R[%v,%v]", r.From, r.To) }

func aToString(a trust.Link) string { return fmt.Sprintf("A[%v,%v]", a.From, a.To) }

func eqToString(expr equations.Expression) string {
	s := &expressionStringer{}
	if err := expr.Accept(s); err != nil {
		return err.Error()
	}
	return s.String()
}

func frtEqToString(eq *equations.Equation) string {
	return rToString(eq.R) + " = " + eqToString(eq.Expression)
}

type expressionStringer struct {
	strings.Builder
}

func (s *expressionStringer) VisitFullUncertainty() (err error) {
	_, err = s.WriteString("U")
	return
}

func (s *expressionStringer) VisitDiscountingRule(r trust.Link, a trust.Link) (err error) {
	_, err = s.WriteString(fmt.Sprintf("(%v ⊠ %v)", rToString(r), aToString(a)))
	return
}

func (s *expressionStringer) VisitDirectReferralTrust(a trust.Link) (err error) {
	_, err = s.WriteString(aToString(a))
	return
}

func (s *expressionStringer) VisitConsensusListStart(count int) (err error) {
	return nil
}

func (s *expressionStringer) VisitConsensusList(index int, equation equations.Expression) (err error) {
	eqStr := eqToString(equation)

	if index == 0 {
		_, err = s.WriteString(eqStr)
	} else {
		if _, err = s.WriteString(" ⊕ "); err != nil {
			return
		}
		_, err = s.WriteString(eqStr)
	}
	return
}

func (s *expressionStringer) VisitConsensusListEnd() (err error) {
	return nil
}
