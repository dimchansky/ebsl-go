package trust

import (
	"fmt"

	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/opinion"
)

// Link represents trust direction
type Link struct {
	From uint64
	To   uint64
}

// String implements fmt.Stringer interface
func (l Link) String() string {
	return fmt.Sprintf("%v -> %v", l.From, l.To)
}

// NextLinkHandler handles next link and returns error
type NextLinkHandler func(Link) error

// LinkIterator used as `foreach` to handle all links
type LinkIterator func(NextLinkHandler) error

// IterableLinks allows to iterate over all links
type IterableLinks interface {
	GetLinkIterator() LinkIterator
}

// NextEvidenceHandler handles next evidence and returns error
type NextEvidenceHandler func(Link, evidence.Type) error

// EvidenceIterator used as `foreach` to handle all evidences
type EvidenceIterator func(NextEvidenceHandler) error

// IterableEvidences allows to iterate over all evidences
type IterableEvidences interface {
	GetEvidenceIterator() EvidenceIterator
}

// DirectReferralEvidence represents direct referral trust matrix in evidence space
type DirectReferralEvidence map[Link]evidence.Type

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

// GetEvidenceIterator implements IterableEvidences interface
func (dre DirectReferralEvidence) GetEvidenceIterator() EvidenceIterator {
	return func(onNext NextEvidenceHandler) error {
		for link, ev := range dre {
			if err := onNext(link, ev); err != nil {
				return err
			}
		}

		return nil
	}
}

// ToDirectReferralOpinion transforms direct referral trust matrix to opinion space
func (dre DirectReferralEvidence) ToDirectReferralOpinion(c uint64) DirectReferralOpinion {
	return make(DirectReferralOpinion, len(dre)).
		FromIterableEvidences(dre, c)
}

// DirectReferralOpinion represents direct referral trust matrix in opinion space
type DirectReferralOpinion map[Link]opinion.Type

// FromIterableEvidences builds DirectReferralOpinion from IterableEvidences
func (dro DirectReferralOpinion) FromIterableEvidences(evidences IterableEvidences, c uint64) DirectReferralOpinion {
	foreachEvidence := evidences.GetEvidenceIterator()
	_ = foreachEvidence(func(link Link, ev evidence.Type) error {
		dro[link] = opinion.FromEvidence(c, ev)
		return nil
	})
	return dro
}

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
type DirectFunctionalTrust map[uint64]opinion.Type

// FinalReferralOpinion represents final referral trust matrix in opinion space
type FinalReferralOpinion map[Link]opinion.Type
