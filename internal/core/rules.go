package core

import "time"

const (
	// RoleAdmin is the system administration role.
	RoleAdmin = "Admin"
	// RoleEditor is the document ingestion and document mutation role.
	RoleEditor = "Editor"
	// RoleReviewer is the review and annotation role.
	RoleReviewer = "Reviewer"

	// StatusDraft is the initial document state before review exposure.
	StatusDraft = "Draft"
	// StatusUnderReview indicates reviewer-visible work in progress.
	StatusUnderReview = "Under Review"
	// StatusRedactionPending indicates staged redactions await confirmation.
	StatusRedactionPending = "Redaction Pending"
	// StatusApproved indicates the document is ready for finalization.
	StatusApproved = "Approved"
	// StatusFinalized is the terminal immutable document state.
	StatusFinalized = "Finalized"
)

// IsValidRole reports whether role is one of the supported local roles.
func IsValidRole(role string) bool {
	return role == RoleAdmin || role == RoleEditor || role == RoleReviewer
}

// CanManageSystem reports whether role may mutate system-level configuration.
func CanManageSystem(role string) bool {
	return role == RoleAdmin
}

// IsValidUserSecret applies the local password complexity floor.
func IsValidUserSecret(secret string) bool {
	if len(secret) < 8 {
		return false
	}
	hasDigit := false
	hasSpecial := false
	for _, r := range secret {
		if r >= '0' && r <= '9' {
			hasDigit = true
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			continue
		}
		hasSpecial = true
	}
	return hasDigit && hasSpecial
}

// CanUploadDocument reports whether role may create new document records.
func CanUploadDocument(role string) bool {
	return role == RoleEditor
}

// CanEditDocument reports whether role has document mutation capability.
func CanEditDocument(role string) bool {
	return role == RoleEditor
}

// CanReviewDocument reports whether role has reviewer-only capabilities.
func CanReviewDocument(role string) bool {
	return role == RoleReviewer
}

// CanReadDocument reports whether role has basic document read capability.
func CanReadDocument(role string) bool {
	return IsValidRole(role)
}

// IsFinalizedStatus reports whether status is the immutable terminal state.
func IsFinalizedStatus(status string) bool {
	return status == StatusFinalized
}

// CanMutateDocument reports whether a document status still allows mutations.
func CanMutateDocument(status string) bool {
	return !IsFinalizedStatus(status)
}

// IsValidDisposition reports whether an annotation disposition is supported.
func IsValidDisposition(v string) bool {
	return v == "Approved" || v == "Rejected" || v == "Needs Discussion"
}

// DefaultDisposition returns the default annotation disposition when omitted.
func DefaultDisposition(v string) string {
	if v == "" {
		return "Needs Discussion"
	}
	return v
}

// IsValidAnnotationType reports whether an annotation type is supported.
func IsValidAnnotationType(v string) bool {
	return v == "Sticky note" || v == "Highlight" || v == "Strikethrough" || v == "Freeform text stamp"
}

// IsValidAnnotationComment enforces the annotation comment length limit.
func IsValidAnnotationComment(comment string) bool {
	return len(comment) <= 2000
}

// IsValidBatesPadding reports whether a Bates padding value is allowed.
func IsValidBatesPadding(padding int) bool {
	return padding >= 0 && padding <= 10
}

// NormalizeBatesStart returns the first valid Bates sequence number.
func NormalizeBatesStart(start int) int {
	if start < 1 {
		return 1
	}
	return start
}

// IsValidRedactionRegion reports whether a redaction region has usable bounds.
func IsValidRedactionRegion(page int, width float64, height float64) bool {
	return page >= 1 && width > 0 && height > 0
}

// IsValidBatchSize reports whether count is inside a positive batch limit.
func IsValidBatchSize(count int, max int) bool {
	return count > 0 && count <= max
}

// IsRequestTimestampFresh reports whether requestTime is inside maxAge of now.
//
// Clock skew is accepted in either direction so clients slightly ahead of the
// server are not rejected as long as the absolute difference is within maxAge.
func IsRequestTimestampFresh(now time.Time, requestTime time.Time, maxAge time.Duration) bool {
	age := now.Sub(requestTime)
	if age < 0 {
		age = -age
	}
	return age <= maxAge
}

// ShouldLockAfterFailedAttempt reports whether the failed-login limit is met.
func ShouldLockAfterFailedAttempt(failedAttemptsAfterIncrement int) bool {
	return failedAttemptsAfterIncrement >= 5
}

// IsAccountLocked reports whether the lockout window is still active.
func IsAccountLocked(now time.Time, lockedUntil *time.Time) bool {
	return lockedUntil != nil && now.Before(*lockedUntil)
}
