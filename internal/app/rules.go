package app

import (
	"time"

	"ironpage-vault/internal/core"
)

func IsValidRole(role string) bool                 { return core.IsValidRole(role) }
func CanManageSystem(role string) bool             { return core.CanManageSystem(role) }
func IsValidUserSecret(secret string) bool         { return core.IsValidUserSecret(secret) }
func CanUploadDocument(role string) bool           { return core.CanUploadDocument(role) }
func CanEditDocument(role string) bool             { return core.CanEditDocument(role) }
func CanReviewDocument(role string) bool           { return core.CanReviewDocument(role) }
func CanReadDocument(role string) bool             { return core.CanReadDocument(role) }
func IsFinalizedStatus(status string) bool         { return core.IsFinalizedStatus(status) }
func CanMutateDocument(status string) bool         { return core.CanMutateDocument(status) }
func IsValidDisposition(v string) bool             { return core.IsValidDisposition(v) }
func DefaultDisposition(v string) string           { return core.DefaultDisposition(v) }
func IsValidAnnotationType(v string) bool          { return core.IsValidAnnotationType(v) }
func IsValidAnnotationComment(comment string) bool { return core.IsValidAnnotationComment(comment) }
func IsValidBatesPadding(padding int) bool         { return core.IsValidBatesPadding(padding) }
func NormalizeBatesStart(start int) int            { return core.NormalizeBatesStart(start) }
func IsValidRedactionRegion(page int, width float64, height float64) bool {
	return core.IsValidRedactionRegion(page, width, height)
}
func IsValidBatchSize(count int, max int) bool { return core.IsValidBatchSize(count, max) }
func IsRequestTimestampFresh(now time.Time, requestTime time.Time, maxAge time.Duration) bool {
	return core.IsRequestTimestampFresh(now, requestTime, maxAge)
}
func ShouldLockAfterFailedAttempt(failedAttemptsAfterIncrement int) bool {
	return core.ShouldLockAfterFailedAttempt(failedAttemptsAfterIncrement)
}
func IsAccountLocked(now time.Time, lockedUntil *time.Time) bool {
	return core.IsAccountLocked(now, lockedUntil)
}
