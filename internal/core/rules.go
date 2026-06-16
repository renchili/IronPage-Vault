package core

import "time"

const (
	RoleAdmin    = "Admin"
	RoleEditor   = "Editor"
	RoleReviewer = "Reviewer"

	StatusDraft            = "Draft"
	StatusUnderReview      = "Under Review"
	StatusRedactionPending = "Redaction Pending"
	StatusApproved         = "Approved"
	StatusFinalized        = "Finalized"
)

func IsValidRole(role string) bool {
	return role == RoleAdmin || role == RoleEditor || role == RoleReviewer
}

func CanManageSystem(role string) bool {
	return role == RoleAdmin
}

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

func CanUploadDocument(role string) bool {
	return role == RoleEditor
}

func CanEditDocument(role string) bool {
	return role == RoleEditor
}

func CanReviewDocument(role string) bool {
	return role == RoleReviewer
}

func CanReadDocument(role string) bool {
	return IsValidRole(role)
}

func IsFinalizedStatus(status string) bool {
	return status == StatusFinalized
}

func CanMutateDocument(status string) bool {
	return !IsFinalizedStatus(status)
}

func IsValidDisposition(v string) bool {
	return v == "Approved" || v == "Rejected" || v == "Needs Discussion"
}

func DefaultDisposition(v string) string {
	if v == "" {
		return "Needs Discussion"
	}
	return v
}

func IsValidAnnotationType(v string) bool {
	return v == "Sticky note" || v == "Highlight" || v == "Strikethrough" || v == "Freeform text stamp"
}

func IsValidAnnotationComment(comment string) bool {
	return len(comment) <= 2000
}

func IsValidBatesPadding(padding int) bool {
	return padding >= 0 && padding <= 10
}

func NormalizeBatesStart(start int) int {
	if start < 1 {
		return 1
	}
	return start
}

func IsValidRedactionRegion(page int, width float64, height float64) bool {
	return page >= 1 && width > 0 && height > 0
}

func IsValidBatchSize(count int, max int) bool {
	return count > 0 && count <= max
}

func IsRequestTimestampFresh(now time.Time, requestTime time.Time, maxAge time.Duration) bool {
	age := now.Sub(requestTime)
	if age < 0 {
		age = -age
	}
	return age <= maxAge
}

func ShouldLockAfterFailedAttempt(failedAttemptsAfterIncrement int) bool {
	return failedAttemptsAfterIncrement >= 5
}

func IsAccountLocked(now time.Time, lockedUntil *time.Time) bool {
	return lockedUntil != nil && now.Before(*lockedUntil)
}
