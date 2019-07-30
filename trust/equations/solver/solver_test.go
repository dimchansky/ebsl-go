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
		{"2",
			trust.DirectReferralEvidence{
				trust.Link{From: 1, To: 2}: evidence.New(400, 300),
				trust.Link{From: 2, To: 3}: evidence.New(10, 5),
				trust.Link{From: 3, To: 4}: evidence.New(500, 0),
				trust.Link{From: 3, To: 5}: evidence.New(500, 0),
				trust.Link{From: 4, To: 5}: evidence.New(500, 0),
				trust.Link{From: 4, To: 6}: evidence.New(500, 0),
				trust.Link{From: 5, To: 6}: evidence.New(500, 0),
				trust.Link{From: 6, To: 7}: evidence.New(5, 5),
			}.ToDirectReferralOpinion(c),
			trust.FinalReferralOpinion{
				trust.Link{From: 1, To: 2}: opinion.New(0.5698005698005698, 0.42735042735042733, 0.002849002849002849),
				trust.Link{From: 1, To: 3}: opinion.New(0.5402485143165856, 0.2701242571582928, 0.18962722852512154),
				trust.Link{From: 1, To: 4}: opinion.New(0.9926504163175847, 0, 0.007349583682415396),
				trust.Link{From: 1, To: 5}: opinion.New(0.9973973565077897, 0, 0.002602643492210305),
				trust.Link{From: 1, To: 6}: opinion.New(0.9979940300054436, 0, 0.0020059699945565415),
				trust.Link{From: 1, To: 7}: opinion.New(0.4165271299394158, 0.4165271299394158, 0.1669457401211684),
				trust.Link{From: 2, To: 3}: opinion.New(0.5882352941176471, 0.29411764705882354, 0.11764705882352941),
				trust.Link{From: 2, To: 4}: opinion.New(0.9932459276916965, 0, 0.006754072308303535),
				trust.Link{From: 2, To: 5}: opinion.New(0.9974771066695858, 0, 0.0025228933304143578),
				trust.Link{From: 2, To: 6}: opinion.New(0.9979947090743448, 0, 0.0020052909256551574),
				trust.Link{From: 2, To: 7}: opinion.New(0.4165271772550089, 0.4165271772550089, 0.16694564548998223),
				trust.Link{From: 3, To: 4}: opinion.New(0.9960159362549801, 0, 0.00398406374501992),
				trust.Link{From: 3, To: 5}: opinion.New(0.998000015936128, 0, 0.001999984063872001),
				trust.Link{From: 3, To: 6}: opinion.New(0.9979980139820138, 0, 0.0020019860179862087),
				trust.Link{From: 3, To: 7}: opinion.New(0.4165274075308264, 0.4165274075308264, 0.16694518493834723),
				trust.Link{From: 4, To: 5}: opinion.New(0.9960159362549801, 0, 0.00398406374501992),
				trust.Link{From: 4, To: 6}: opinion.New(0.998000015936128, 0, 0.001999984063872001),
				trust.Link{From: 4, To: 7}: opinion.New(0.4165275470202235, 0.4165275470202235, 0.16694490595955308),
				trust.Link{From: 5, To: 6}: opinion.New(0.9960159362549801, 0, 0.00398406374501992),
				trust.Link{From: 5, To: 7}: opinion.New(0.4163890739506996, 0.4163890739506996, 0.1672218520986009),
				trust.Link{From: 6, To: 7}: opinion.New(0.4166666666666667, 0.4166666666666667, 0.16666666666666666),
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
