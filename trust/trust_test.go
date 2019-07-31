package trust_test

import (
	"fmt"
	"testing"

	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/opinion"
	"github.com/dimchansky/ebsl-go/trust"
	"github.com/dimchansky/ebsl-go/trust/equations"
	"github.com/dimchansky/ebsl-go/trust/equations/solver"
)

func BenchmarkCreateEquationsAndSolveThemInOneEpoch(b *testing.B) {
	for _, nodes := range []uint64{
		10,
		20,
		50,
		100,
	} {
		b.Run(fmt.Sprintf("%v nodes", nodes), func(b *testing.B) {
			dro := make(trust.DirectReferralOpinion, nodes*(nodes-1))
			for i := uint64(1); i <= nodes; i++ {
				for j := uint64(1); j <= nodes; j++ {
					if i != j {
						dro[trust.Link{From: i, To: j}] = opinion.FromEvidence(2, evidence.New(1, 1))
					}
				}
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				eqs := equations.CreateFinalReferralTrustEquations(dro)
				context := equations.NewDefaultFinalReferralTrustEquationContext(dro)

				if err := solver.SolveFinalReferralTrustEquations(
					context,
					eqs,
					solver.UseMaxEpochs(1),
				); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
