package app

import (
	"ironpage-vault/internal/core"
	"ironpage-vault/internal/store"
)

func corePrincipal(p Principal) core.Principal {
	return core.Principal{UserID: p.UserID, Role: p.Role}
}

func coreDocumentAccess(d Document) core.DocumentAccess {
	return core.DocumentAccess{OwnerID: d.OwnerID, Status: d.Status}
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
	return store.DocumentListWhereClause(store.PrincipalFilter{UserID: p.UserID, Role: p.Role})
}
