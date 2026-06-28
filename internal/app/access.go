// Package app contains Echo routing, middleware, request binding, and response mapping.
//
// This layer adapts HTTP requests to application behavior and coordinates core,
// store, service, and platform packages. Domain policy should move downward into
// core or service packages when it does not require HTTP-specific state.
package app

import (
	"ironpage-vault/internal/core"
	"ironpage-vault/internal/store"
)

// corePrincipal adapts the app-layer Principal to the infrastructure-free core type.
func corePrincipal(p Principal) core.Principal {
	return core.Principal{UserID: p.UserID, Role: p.Role}
}

// coreDocumentAccess adapts app document records to the core policy input.
func coreDocumentAccess(d Document) core.DocumentAccess {
	return core.DocumentAccess{OwnerID: d.OwnerID, Status: d.Status}
}

// canReadDocumentObject keeps handlers on the shared core object-access rule.
func canReadDocumentObject(p Principal, d Document) bool {
	return core.CanReadDocumentObject(corePrincipal(p), coreDocumentAccess(d))
}

// canEditDocumentObject keeps edit checks aligned with core immutability rules.
func canEditDocumentObject(p Principal, d Document) bool {
	return core.CanEditDocumentObject(corePrincipal(p), coreDocumentAccess(d))
}

// canReviewDocumentObject keeps review checks aligned with core workflow rules.
func canReviewDocumentObject(p Principal, d Document) bool {
	return core.CanReviewDocumentObject(corePrincipal(p), coreDocumentAccess(d))
}

// canTransitionDocumentObject keeps workflow checks aligned with core policy.
func canTransitionDocumentObject(p Principal, d Document) bool {
	return core.CanTransitionDocumentObject(corePrincipal(p), coreDocumentAccess(d))
}

// documentListWhereClause delegates SQL filtering to store so app owns policy use
// while store owns SQL shape.
func documentListWhereClause(p Principal) (string, []interface{}) {
	return store.DocumentListWhereClause(store.PrincipalFilter{UserID: p.UserID, Role: p.Role})
}
