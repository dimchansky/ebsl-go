package trust

import (
	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/opinion"
)

type Link struct {
	From uint64
	To   uint64
}

type LinkHandler func(Link) error

type LinkIterator func(LinkHandler) error

type IterableLinks interface {
	GetLinkIterator() LinkIterator
}

type DirectReferralEvidence map[Link]*evidence.Type

func (dre DirectReferralEvidence) GetLinkIterator() LinkIterator {
	return func(onNext LinkHandler) error {
		for link := range dre {
			if err := onNext(link); err != nil {
				return err
			}
		}

		return nil
	}
}

func (dre DirectReferralEvidence) ToDirectReferralOpinion(c uint64) DirectReferralOpinion {
	res := make(DirectReferralOpinion, len(dre))
	for ref, ev := range dre {
		res[ref] = opinion.FromEvidence(c, ev)
	}
	return res
}

type DirectReferralOpinion map[Link]*opinion.Type

func (dro DirectReferralOpinion) GetLinkIterator() LinkIterator {
	return func(onNext LinkHandler) error {
		for link := range dro {
			if err := onNext(link); err != nil {
				return err
			}
		}

		return nil
	}
}

type FinalReferralTrust map[Link]*opinion.Type
