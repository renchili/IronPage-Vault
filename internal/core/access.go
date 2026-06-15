package core

type Principal struct {
    UserID string
    Role   string
}

type DocumentAccess struct {
    OwnerID string
    Status  string
}

func CanReadDocumentObject(p Principal, d DocumentAccess) bool {
    if p.Role == RoleAdmin { return true }
    if d.OwnerID == p.UserID { return true }
    if p.Role == RoleReviewer { return d.Status != StatusDraft }
    return false
}

func CanEditDocumentObject(p Principal, d DocumentAccess) bool {
    return p.Role == RoleEditor && d.OwnerID == p.UserID && d.Status != StatusFinalized
}

func CanReviewDocumentObject(p Principal, d DocumentAccess) bool {
    return p.Role == RoleReviewer && d.Status != StatusDraft && d.Status != StatusFinalized
}

func CanTransitionDocumentObject(p Principal, d DocumentAccess) bool {
    if d.Status == StatusFinalized { return false }
    if p.Role == RoleEditor && d.OwnerID == p.UserID { return true }
    if p.Role == RoleReviewer && d.Status != StatusDraft { return true }
    return false
}
