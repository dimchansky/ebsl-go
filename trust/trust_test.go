package trust_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/trust"
)

type expressionStringer struct {
	strings.Builder
}

func (s *expressionStringer) VisitFullUncertainty() {
	_, _ = s.WriteString("U")
}

func (s *expressionStringer) VisitDiscountingRule(r trust.Link, a trust.Link) {
	_, _ = s.WriteString(fmt.Sprintf("(%v ⊠ %v)", rToString(r), aToString(a)))
}

func (s *expressionStringer) VisitDirectReferralTrust(a trust.Link) {
	_, _ = s.WriteString(aToString(a))
}

func (s *expressionStringer) VisitConsensusListStart(count int) {
}

func (s *expressionStringer) VisitConsensusList(index int, equation trust.Expression) {
	eqStr := eqToString(equation)

	if index == 0 {
		_, _ = s.WriteString(eqStr)
	} else {
		_, _ = s.WriteString(" ⊕ ")
		_, _ = s.WriteString(eqStr)
	}
}

func (s *expressionStringer) VisitConsensusListEnd() {
}

func rToString(r trust.Link) string { return fmt.Sprintf("R[%v,%v]", r.From, r.To) }
func aToString(a trust.Link) string { return fmt.Sprintf("A[%v,%v]", a.From, a.To) }
func eqToString(expr trust.Expression) string {
	s := &expressionStringer{}
	expr.Accept(s)
	return s.String()
}
func frtEqToString(eq *trust.FinalReferralTrustEquation) string {
	return rToString(eq.R) + " = " + eqToString(eq.Expression)
}

func logEquations(t *testing.T, eqs []*trust.FinalReferralTrustEquation) {
	for _, eq := range eqs {
		t.Log(frtEqToString(eq))
	}
}

func TestExample(t *testing.T) {
	c := uint64(2)

	a := trust.DirectReferralEvidence{
		trust.Link{From: 1, To: 2}: evidence.New(1, 1),
		trust.Link{From: 2, To: 3}: evidence.New(1, 1),
		trust.Link{From: 3, To: 2}: evidence.New(1, 1),
	}.ToDirectReferralOpinion(c)

	logEquations(t, trust.CreateFinalReferralTrustEquations(a))
}

func TestExample2(t *testing.T) {
	c := uint64(2)

	a := trust.DirectReferralEvidence{
		trust.Link{From: 1, To: 2}: evidence.New(1, 1),
		trust.Link{From: 2, To: 3}: evidence.New(1, 1),
		trust.Link{From: 3, To: 4}: evidence.New(1, 1),
		trust.Link{From: 3, To: 5}: evidence.New(1, 1),
		trust.Link{From: 4, To: 5}: evidence.New(1, 1),
		trust.Link{From: 4, To: 6}: evidence.New(1, 1),
		trust.Link{From: 5, To: 6}: evidence.New(1, 1),
		trust.Link{From: 6, To: 7}: evidence.New(1, 1),
	}.ToDirectReferralOpinion(c)

	logEquations(t, trust.CreateFinalReferralTrustEquations(a))
}
