package trust

import (
	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/opinion"
)

type Link struct {
	From uint64
	To   uint64
}

type DirectReferralEvidence map[Link]*evidence.Type

func (dre DirectReferralEvidence) ToDirectReferralOpinion(c uint64) DirectReferralOpinion {
	res := make(DirectReferralOpinion, len(dre))
	for ref, ev := range dre {
		res[ref] = opinion.FromEvidence(c, ev)
	}
	return res
}

type DirectReferralOpinion map[Link]*opinion.Type

type FinalReferralTrust map[Link]*opinion.Type
