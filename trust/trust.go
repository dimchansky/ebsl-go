package trust

import (
	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/opinion"
)

// Link represents trust direction
type Link struct {
	From uint64
	To   uint64
}

// NextLinkHandler handles next link and returns error
type NextLinkHandler func(Link) error

// LinkIterator used as `foreach` to handle all links
type LinkIterator func(NextLinkHandler) error

// IterableLinks allows to iterate over all links
type IterableLinks interface {
	GetLinkIterator() LinkIterator
}

// DirectReferralEvidence represents direct referral trust matrix in evidence space
type DirectReferralEvidence map[Link]*evidence.Type

// GetLinkIterator implements IterableLinks interface
func (dre DirectReferralEvidence) GetLinkIterator() LinkIterator {
	return func(onNext NextLinkHandler) error {
		for link := range dre {
			if err := onNext(link); err != nil {
				return err
			}
		}

		return nil
	}
}

// ToDirectReferralOpinion transforms direct referral trust matrix to opinion space
func (dre DirectReferralEvidence) ToDirectReferralOpinion(c uint64) DirectReferralOpinion {
	res := make(DirectReferralOpinion, len(dre))
	for ref, ev := range dre {
		res[ref] = opinion.FromEvidence(c, ev)
	}
	return res
}

// DirectReferralOpinion represents direct referral trust matrix in opinion space
type DirectReferralOpinion map[Link]*opinion.Type

// GetLinkIterator implements IterableLinks interface
func (dro DirectReferralOpinion) GetLinkIterator() LinkIterator {
	return func(onNext NextLinkHandler) error {
		for link := range dro {
			if err := onNext(link); err != nil {
				return err
			}
		}

		return nil
	}
}

// DirectFunctionalTrust is the direct opinion about an entity's ability to provide a specific function
type DirectFunctionalTrust map[uint64]*opinion.Type

// FinalReferralOpinion represents final referral trust matrix in opinion space
type FinalReferralOpinion map[Link]*opinion.Type
