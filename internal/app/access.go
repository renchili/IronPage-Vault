package app

import "ironpage-vault/internal/core"

func corePrincipal(p Principal) core.Principal {
    return core.Principal{UserID:p.UserID, Role:p.Role}
}

func coreDocumentAccess(d Document) core.DocumentAccess {
    return core.DocumentAccess{OwnerID:d.OwnerID, Status:d.Status}
}

func canReadDocumentObject(p Principal, d Document) bool {
    return core.CanReadDocumentObject(corePrincipal(p), coreDocumentAccess(d))
}

func canEditDocumentObject(p Principal, d Document) bool {
    return core.CanEditDocumentObject(corePrincipal(p), coreDocumentAccess(d))
}

func canReviewDocumentObject(p Principal, d Document) bool {
    return core.CanReviewDocumentObject(corePrincipal(p), coreDocumentAccess(d))
}

func canTransitionDocumentObject(p Principal, d Document) bool {
    return core.CanTransitionDocumentObject(corePrincipal(p), coreDocumentAccess(d))
}

func documentListWhereClause(p Principal) (string, []interface{}) {
    if p.Role == RoleAdmin { return "1=1", nil }
    if p.Role == RoleReviewer { return "status <> 'Draft'", nil }
    return "owner_id=$1", []interface{}{p.UserID}
}
