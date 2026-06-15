package core

import (
    "testing"
    "time"
)

func TestRoleRules(t *testing.T) {
    for _, role := range []string{RoleAdmin, RoleEditor, RoleReviewer} {
        if !IsValidRole(role) { t.Fatalf("expected valid role %s", role) }
        if !CanReadDocument(role) { t.Fatalf("expected readable role %s", role) }
    }
    if IsValidRole("Owner") { t.Fatalf("unexpected valid role") }
    if !CanManageSystem(RoleAdmin) || CanManageSystem(RoleEditor) || CanManageSystem(RoleReviewer) { t.Fatalf("manage-system rule mismatch") }
    if !CanUploadDocument(RoleEditor) || CanUploadDocument(RoleAdmin) || CanUploadDocument(RoleReviewer) { t.Fatalf("upload rule mismatch") }
    if !CanReviewDocument(RoleReviewer) || CanReviewDocument(RoleEditor) || CanReviewDocument(RoleAdmin) { t.Fatalf("review rule mismatch") }
}

func TestDocumentAndValidationRules(t *testing.T) {
    if !IsFinalizedStatus(StatusFinalized) { t.Fatalf("finalized status mismatch") }
    if CanMutateDocument(StatusFinalized) { t.Fatalf("finalized document should not mutate") }
    if !CanMutateDocument(StatusDraft) { t.Fatalf("draft should mutate") }
    if !IsValidDisposition("Approved") || IsValidDisposition("Pending") { t.Fatalf("disposition rule mismatch") }
    if DefaultDisposition("") != "Needs Discussion" { t.Fatalf("default disposition mismatch") }
    if !IsValidAnnotationType("Highlight") || IsValidAnnotationType("Circle") { t.Fatalf("annotation type rule mismatch") }
    if !IsValidAnnotationComment(string(make([]byte, 2000))) || IsValidAnnotationComment(string(make([]byte, 2001))) { t.Fatalf("annotation comment rule mismatch") }
    if !IsValidBatesPadding(0) || !IsValidBatesPadding(10) || IsValidBatesPadding(11) { t.Fatalf("Bates padding rule mismatch") }
    if NormalizeBatesStart(0) != 1 || NormalizeBatesStart(9) != 9 { t.Fatalf("Bates start normalization mismatch") }
    if !IsValidRedactionRegion(1, 1, 1) || IsValidRedactionRegion(0, 1, 1) { t.Fatalf("redaction region rule mismatch") }
    if !IsValidBatchSize(1, 250) || IsValidBatchSize(251, 250) { t.Fatalf("batch size rule mismatch") }
}

func TestFreshnessAndLockoutRules(t *testing.T) {
    now := time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC)
    if !IsRequestTimestampFresh(now, now.Add(-30*time.Second), time.Minute) { t.Fatalf("fresh timestamp rejected") }
    if IsRequestTimestampFresh(now, now.Add(-2*time.Minute), time.Minute) { t.Fatalf("stale timestamp accepted") }
    if ShouldLockAfterFailedAttempt(4) || !ShouldLockAfterFailedAttempt(5) { t.Fatalf("lockout threshold mismatch") }
    until := now.Add(time.Minute)
    if !IsAccountLocked(now, &until) { t.Fatalf("expected account locked") }
    past := now.Add(-time.Minute)
    if IsAccountLocked(now, &past) || IsAccountLocked(now, nil) { t.Fatalf("account lock mismatch") }
}
