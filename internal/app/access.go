package app

func canReadDocumentObject(p Principal, d Document) bool {
    if p.Role == RoleAdmin { return true }
    if d.OwnerID == p.UserID { return true }
    if p.Role == RoleReviewer { return d.Status != StatusDraft }
    return false
}

func documentListWhereClause(p Principal) (string, []interface{}) {
    if p.Role == RoleAdmin { return "1=1", nil }
    if p.Role == RoleReviewer { return "status <> 'Draft'", nil }
    return "owner_id=$1", []interface{}{p.UserID}
}
