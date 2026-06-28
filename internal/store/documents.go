package store

const (
	// RoleAdmin is the store-layer role value used for SQL visibility filters.
	RoleAdmin = "Admin"
	// RoleReviewer is the store-layer role value used for review visibility filters.
	RoleReviewer = "Reviewer"
)

// PrincipalFilter carries only the caller fields required for SQL filtering.
type PrincipalFilter struct {
	UserID string
	Role   string
}

// DocumentListWhereClause returns the SQL predicate for document list visibility.
//
// The app layer decides when this policy is used; the store layer owns the SQL
// shape so handlers do not assemble query fragments directly.
func DocumentListWhereClause(p PrincipalFilter) (string, []interface{}) {
	if p.Role == RoleAdmin {
		return "1=1", nil
	}
	if p.Role == RoleReviewer {
		return "status <> 'Draft'", nil
	}
	return "owner_id=$1", []interface{}{p.UserID}
}
