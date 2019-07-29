package equations_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/dimchansky/ebsl-go/trust"
	"github.com/dimchansky/ebsl-go/trust/equations"
	"github.com/go-test/deep"
)

func TestCreateFinalReferralTrustEquations(t *testing.T) {
	tests := []struct {
		name  string
		links links
		want  strEquations
	}{
		{"1",
			links{
				trust.Link{From: 1, To: 2},
				trust.Link{From: 2, To: 3},
				trust.Link{From: 3, To: 2},
			},
			strEquations{
				trust.Link{From: 1, To: 2}: "(R[1,3] ⊠ A[3,2]) ⊕ A[1,2]",
				trust.Link{From: 1, To: 3}: "(R[1,2] ⊠ A[2,3])",
				trust.Link{From: 2, To: 3}: "A[2,3]",
				trust.Link{From: 3, To: 2}: "A[3,2]",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toStringEquations(equations.CreateFinalReferralTrustEquations(tt.links))

			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("CreateFinalReferralTrustEquations: %v", diff)
			}
		})
	}
}

type links []trust.Link

func (l links) GetLinkIterator() trust.LinkIterator {
	return func(onNext trust.NextLinkHandler) error {
		for _, value := range l {
			if err := onNext(value); err != nil {
				return err
			}
		}
		return nil
	}
}

type strEquations map[trust.Link]string

func toStringEquations(eqs equations.FinalReferralTrustEquations) strEquations {
	r := make(strEquations, len(eqs))
	for _, eq := range eqs {
		r[eq.R] = expressionToString(eq.Expression)
	}
	return r
}

func expressionToString(expr equations.FinalReferralTrustExpression) string {
	s := &expressionStringer{}
	if err := expr.Accept(s); err != nil {
		return err.Error()
	}
	return s.String()
}

type expressionStringer struct {
	strings.Builder
	consensusList []string
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
	s.consensusList = nil
	return nil
}

func (s *expressionStringer) VisitConsensusList(index int, equation equations.FinalReferralTrustExpression) (err error) {
	eqStr := expressionToString(equation)

	s.consensusList = append(s.consensusList, eqStr)

	return
}

func (s *expressionStringer) VisitConsensusListEnd() (err error) {
	sort.Strings(s.consensusList) // sorting consensus list to have consistent result
	_, err = s.WriteString(strings.Join(s.consensusList, " ⊕ "))
	return
}

func rToString(r trust.Link) string { return fmt.Sprintf("R[%v,%v]", r.From, r.To) }

func aToString(a trust.Link) string { return fmt.Sprintf("A[%v,%v]", a.From, a.To) }