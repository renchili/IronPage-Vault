package app

import (
	"time"

	"ironpage-vault/internal/core"
)

// IsValidRole delegates to core while app handlers are migrated to core rules.
func IsValidRole(role string) bool { return core.IsValidRole(role) }

// CanManageSystem delegates system-management capability checks to core.
func CanManageSystem(role string) bool { return core.CanManageSystem(role) }

// IsValidUserSecret delegates local secret validation to core.
func IsValidUserSecret(secret string) bool { return core.IsValidUserSecret(secret) }

// CanUploadDocument delegates upload capability checks to core.
func CanUploadDocument(role string) bool { return core.CanUploadDocument(role) }

// CanEditDocument delegates document edit capability checks to core.
func CanEditDocument(role string) bool { return core.CanEditDocument(role) }

// CanReviewDocument delegates review capability checks to core.
func CanReviewDocument(role string) bool { return core.CanReviewDocument(role) }

// CanReadDocument delegates broad document read capability checks to core.
func CanReadDocument(role string) bool { return core.CanReadDocument(role) }

// IsFinalizedStatus delegates terminal-state detection to core.
func IsFinalizedStatus(status string) bool { return core.IsFinalizedStatus(status) }

// CanMutateDocument delegates finalized immutability checks to core.
func CanMutateDocument(status string) bool { return core.CanMutateDocument(status) }

// IsValidDisposition delegates annotation disposition validation to core.
func IsValidDisposition(v string) bool { return core.IsValidDisposition(v) }

// DefaultDisposition delegates annotation disposition defaulting to core.
func DefaultDisposition(v string) string { return core.DefaultDisposition(v) }

// IsValidAnnotationType delegates annotation type validation to core.
func IsValidAnnotationType(v string) bool { return core.IsValidAnnotationType(v) }

// IsValidAnnotationComment delegates annotation comment length checks to core.
func IsValidAnnotationComment(comment string) bool { return core.IsValidAnnotationComment(comment) }

// IsValidBatesPadding delegates Bates padding validation to core.
func IsValidBatesPadding(padding int) bool { return core.IsValidBatesPadding(padding) }

// NormalizeBatesStart delegates Bates sequence normalization to core.
func NormalizeBatesStart(start int) int { return core.NormalizeBatesStart(start) }

// IsValidRedactionRegion delegates redaction geometry validation to core.
func IsValidRedactionRegion(page int, width float64, height float64) bool {
	return core.IsValidRedactionRegion(page, width, height)
}

// IsValidBatchSize delegates batch size validation to core.
func IsValidBatchSize(count int, max int) bool { return core.IsValidBatchSize(count, max) }

// IsRequestTimestampFresh delegates request freshness validation to core.
func IsRequestTimestampFresh(now time.Time, requestTime time.Time, maxAge time.Duration) bool {
	return core.IsRequestTimestampFresh(now, requestTime, maxAge)
}

// ShouldLockAfterFailedAttempt delegates login lockout threshold checks to core.
func ShouldLockAfterFailedAttempt(failedAttemptsAfterIncrement int) bool {
	return core.ShouldLockAfterFailedAttempt(failedAttemptsAfterIncrement)
}

// IsAccountLocked delegates active lockout-window checks to core.
func IsAccountLocked(now time.Time, lockedUntil *time.Time) bool {
	return core.IsAccountLocked(now, lockedUntil)
}
