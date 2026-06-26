package core

// Principal is the minimal caller identity used by core object-level policy.
//
// The core package intentionally keeps this type independent of HTTP sessions or
// database rows so authorization rules can be tested without infrastructure.
type Principal struct {
	UserID string
	Role   string
}

// DocumentAccess contains only the document attributes needed for access checks.
type DocumentAccess struct {
	OwnerID string
	Status  string
}

// CanReadDocumentObject reports whether a caller may read a specific document.
//
// Admins can read for oversight, owners can read their own documents, and
// reviewers can read documents after Draft so review access does not leak drafts.
func CanReadDocumentObject(p Principal, d DocumentAccess) bool {
	if p.Role == RoleAdmin {
		return true
	}
	if d.OwnerID == p.UserID {
		return true
	}
	if p.Role == RoleReviewer {
		return d.Status != StatusDraft
	}
	return false
}

// CanEditDocumentObject reports whether a caller may mutate document content.
//
// Editing is limited to the owning Editor and always blocked after Finalized.
func CanEditDocumentObject(p Principal, d DocumentAccess) bool {
	return p.Role == RoleEditor && d.OwnerID == p.UserID && d.Status != StatusFinalized
}

// CanReviewDocumentObject reports whether a Reviewer may add review activity.
func CanReviewDocumentObject(p Principal, d DocumentAccess) bool {
	return p.Role == RoleReviewer && d.Status != StatusDraft && d.Status != StatusFinalized
}

// CanTransitionDocumentObject reports whether a caller may move workflow state.
//
// Finalized is terminal. Editors can transition documents they own, while
// reviewers can move non-draft documents through review-side states.
func CanTransitionDocumentObject(p Principal, d DocumentAccess) bool {
	if d.Status == StatusFinalized {
		return false
	}
	if p.Role == RoleEditor && d.OwnerID == p.UserID {
		return true
	}
	if p.Role == RoleReviewer && d.Status != StatusDraft {
		return true
	}
	return false
}
