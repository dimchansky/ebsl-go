package trust_test

import (
	"testing"

	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/trust"
)

func Test(t *testing.T) {
	c := uint64(2)

	a := trust.DirectReferralEvidence{
		trust.Link{From: 1, To: 2}: evidence.New(1, 1),
		trust.Link{From: 2, To: 3}: evidence.New(1, 1),
		trust.Link{From: 3, To: 2}: evidence.New(1, 1),
	}.ToDirectReferralOpinion(c)

	a.XXX()
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

	a.XXX()
}
