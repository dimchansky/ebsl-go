package trust_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/trust"
)

type equationStringer struct {
	strings.Builder
}

func (s *equationStringer) VisitFullUncertainty() {
	_, _ = s.WriteString("U")
}

func (s *equationStringer) VisitDiscountingRule(r trust.R, a trust.A) {
	_, _ = s.WriteString(fmt.Sprintf("(%v ⊠ %v)", rToString(r), aToString(a)))
}

func (s *equationStringer) VisitDirectReferralTrust(a trust.A) {
	_, _ = s.WriteString(aToString(a))
}

func (s *equationStringer) VisitConsensusStart() {}

func (s *equationStringer) VisitConsensusOpinion(index int, equation trust.Equation) {
	eqStr := eqToString(equation)

	if index == 0 {
		_, _ = s.WriteString(eqStr)
	} else {
		_, _ = s.WriteString(" ⊕ ")
		_, _ = s.WriteString(eqStr)
	}
}

func (s *equationStringer) VisitConsensusEnd() {}

func rToString(r trust.R) string { return fmt.Sprintf("R[%v,%v]", r.From, r.To) }
func aToString(a trust.A) string { return fmt.Sprintf("A[%v,%v]", a.From, a.To) }
func eqToString(eq trust.Equation) string {
	s := &equationStringer{}
	eq.Accept(s)
	return s.String()
}

func logEquations(t *testing.T, eqs []*trust.FinalReferralTrustEquation) {
	for _, eq := range eqs {
		t.Log(rToString(eq.R), "=", eqToString(eq.Equation))
	}
}

func Test(t *testing.T) {
	c := uint64(2)

	a := trust.DirectReferralEvidence{
		trust.Link{From: 1, To: 2}: evidence.New(1, 1),
		trust.Link{From: 2, To: 3}: evidence.New(1, 1),
		trust.Link{From: 3, To: 2}: evidence.New(1, 1),
	}.ToDirectReferralOpinion(c)

	logEquations(t, a.CreateFinalReferralTrustEquations())
}

func Test2(t *testing.T) {
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

	logEquations(t, a.CreateFinalReferralTrustEquations())
}
