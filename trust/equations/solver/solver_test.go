package solver_test

import (
	"testing"

	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/opinion"
	"github.com/dimchansky/ebsl-go/trust"
	"github.com/dimchansky/ebsl-go/trust/equations"
	"github.com/dimchansky/ebsl-go/trust/equations/solver"
	"github.com/go-test/deep"
)

func TestSolveFinalReferralTrustEquations(t *testing.T) {
	c := uint64(2)

	tests := []struct {
		name string
		dro  trust.DirectReferralOpinion
		want trust.FinalReferralOpinion
	}{
		{"1",
			trust.DirectReferralEvidence{
				trust.Link{From: 1, To: 2}: evidence.New(2, 2),
				trust.Link{From: 2, To: 3}: evidence.New(2, 2),
				trust.Link{From: 3, To: 2}: evidence.New(2, 2),
			}.ToDirectReferralOpinion(c),
			trust.FinalReferralOpinion{
				trust.Link{From: 1, To: 2}: opinion.New(0.3535533905932738, 0.3535533905932738, 0.2928932188134525),
				trust.Link{From: 1, To: 3}: opinion.New(0.20710678118654752, 0.20710678118654752, 0.585786437626905),
				trust.Link{From: 2, To: 3}: opinion.New(0.3333333333333333, 0.3333333333333333, 0.3333333333333333),
				trust.Link{From: 3, To: 2}: opinion.New(0.3333333333333333, 0.3333333333333333, 0.3333333333333333),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eqs := equations.CreateFinalReferralTrustEquations(tt.dro)

			context := equations.NewDefaultFinalReferralTrustEquationContext(tt.dro)

			if err := solver.SolveFinalReferralTrustEquations(
				context,
				eqs,
				solver.UseOnEpochEndCallback(func(epoch uint, aggregatedDistance float64) error {
					t.Logf("Epoch %v error: %v\n", epoch, aggregatedDistance)
					return nil
				}),
			); err != nil {
				t.Fatal(err)
			}

			got := context.FinalReferralTrust

			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("SolveFinalReferralTrustEquations: %v", diff)
			}
		})
	}
}
