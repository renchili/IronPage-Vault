package app

import (
	"errors"
	"testing"
)

func TestVersionLimitAllowsFortyNineToFifty(t *testing.T) {
	next, err := nextDocumentVersion(49, 50)
	if err != nil {
		t.Fatalf("49 to 50 was rejected: %v", err)
	}
	if next != 50 {
		t.Fatalf("next version = %d, want 50", next)
	}
}

func TestVersionLimitRejectsFiftyToFiftyOne(t *testing.T) {
	if _, err := nextDocumentVersion(50, 50); !errors.Is(err, errVersionLimitReached) {
		t.Fatalf("50 to 51 error = %v, want version limit", err)
	}
}

func TestRollbackVersionMustRemainWithinCeiling(t *testing.T) {
	for _, version := range []int{1, 49, 50} {
		if !validRollbackVersion(version, 50) {
			t.Fatalf("valid rollback target %d was rejected", version)
		}
	}
	for _, version := range []int{0, 51} {
		if validRollbackVersion(version, 50) {
			t.Fatalf("invalid rollback target %d was accepted", version)
		}
	}
}
