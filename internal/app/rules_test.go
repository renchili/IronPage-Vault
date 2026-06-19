package app

import (
	"testing"
	"time"
)

func TestRoleRules(t *testing.T) {
	valid := []string{RoleAdmin, RoleEditor, RoleReviewer}
	for _, role := range valid {
		if !IsValidRole(role) {
			t.Fatalf("expected role %s to be valid", role)
		}
		if !CanReadDocument(role) {
			t.Fatalf("expected role %s to read documents", role)
		}
	}
	if IsValidRole("Owner") {
		t.Fatalf("unexpected valid role")
	}
	if !CanManageSystem(RoleAdmin) || CanManageSystem(RoleEditor) || CanManageSystem(RoleReviewer) {
		t.Fatalf("system management role rule mismatch")
	}
	if !CanUploadDocument(RoleEditor) || CanUploadDocument(RoleAdmin) || CanUploadDocument(RoleReviewer) {
		t.Fatalf("upload role rule mismatch")
	}
	if !CanReviewDocument(RoleReviewer) || CanReviewDocument(RoleEditor) || CanReviewDocument(RoleAdmin) {
		t.Fatalf("review role rule mismatch")
	}
}

func TestDocumentMutationRules(t *testing.T) {
	if !IsFinalizedStatus(StatusFinalized) {
		t.Fatalf("Finalized must be finalized")
	}
	if CanMutateDocument(StatusFinalized) {
		t.Fatalf("Finalized document must not be mutable")
	}
	if !CanMutateDocument(StatusDraft) {
		t.Fatalf("Draft document should be mutable")
	}
}

func TestAnnotationRules(t *testing.T) {
	validTypes := []string{"Sticky note", "Highlight", "Strikethrough", "Freeform text stamp"}
	for _, v := range validTypes {
		if !IsValidAnnotationType(v) {
			t.Fatalf("expected annotation type %q to be valid", v)
		}
	}
	if IsValidAnnotationType("Circle") {
		t.Fatalf("unexpected annotation type accepted")
	}
	if !IsValidAnnotationComment(string(make([]byte, 2000))) {
		t.Fatalf("2000 byte comment should be valid")
	}
	if IsValidAnnotationComment(string(make([]byte, 2001))) {
		t.Fatalf("2001 byte comment should be invalid")
	}
}

func TestDispositionRules(t *testing.T) {
	for _, v := range []string{"Approved", "Rejected", "Needs Discussion"} {
		if !IsValidDisposition(v) {
			t.Fatalf("expected disposition %q to be valid", v)
		}
	}
	if IsValidDisposition("Pending") {
		t.Fatalf("unexpected disposition accepted")
	}
	if got := DefaultDisposition(""); got != "Needs Discussion" {
		t.Fatalf("default disposition=%q", got)
	}
	if got := DefaultDisposition("Approved"); got != "Approved" {
		t.Fatalf("explicit disposition changed to %q", got)
	}
}

func TestBatesRules(t *testing.T) {
	for _, padding := range []int{0, 1, 10} {
		if !IsValidBatesPadding(padding) {
			t.Fatalf("padding %d should be valid", padding)
		}
	}
	for _, padding := range []int{-1, 11} {
		if IsValidBatesPadding(padding) {
			t.Fatalf("padding %d should be invalid", padding)
		}
	}
	if NormalizeBatesStart(0) != 1 || NormalizeBatesStart(-4) != 1 || NormalizeBatesStart(7) != 7 {
		t.Fatalf("Bates start normalization mismatch")
	}
}

func TestRedactionAndBatchRules(t *testing.T) {
	if !IsValidRedactionRegion(1, 10, 20) {
		t.Fatalf("valid redaction rejected")
	}
	if IsValidRedactionRegion(0, 10, 20) || IsValidRedactionRegion(1, 0, 20) || IsValidRedactionRegion(1, 10, 0) {
		t.Fatalf("invalid redaction accepted")
	}
	if !IsValidBatchSize(1, 250) || !IsValidBatchSize(250, 250) {
		t.Fatalf("valid batch size rejected")
	}
	if IsValidBatchSize(0, 250) || IsValidBatchSize(251, 250) {
		t.Fatalf("invalid batch size accepted")
	}
}

func TestRequestFreshnessAndLockoutRules(t *testing.T) {
	now := time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC)
	if !IsRequestTimestampFresh(now, now.Add(-30*time.Second), time.Minute) {
		t.Fatalf("fresh past timestamp rejected")
	}
	if !IsRequestTimestampFresh(now, now.Add(30*time.Second), time.Minute) {
		t.Fatalf("fresh future timestamp rejected")
	}
	if IsRequestTimestampFresh(now, now.Add(-2*time.Minute), time.Minute) {
		t.Fatalf("stale timestamp accepted")
	}
	if ShouldLockAfterFailedAttempt(4) || !ShouldLockAfterFailedAttempt(5) {
		t.Fatalf("lockout threshold mismatch")
	}
	until := now.Add(time.Minute)
	if !IsAccountLocked(now, &until) {
		t.Fatalf("account should be locked")
	}
	past := now.Add(-time.Minute)
	if IsAccountLocked(now, &past) || IsAccountLocked(now, nil) {
		t.Fatalf("account lock rule mismatch")
	}
}
