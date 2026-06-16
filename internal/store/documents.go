package store

const (
    RoleAdmin    = "Admin"
    RoleReviewer = "Reviewer"
)

type PrincipalFilter struct {
    UserID string
    Role   string
}

func DocumentListWhereClause(p PrincipalFilter) (string, []interface{}) {
    if p.Role == RoleAdmin { return "1=1", nil }
    if p.Role == RoleReviewer { return "status <> 'Draft'", nil }
    return "owner_id=$1", []interface{}{p.UserID}
}
