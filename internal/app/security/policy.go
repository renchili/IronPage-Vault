package security

const (
	RoleAdmin    = "Admin"
	RoleEditor   = "Editor"
	RoleReviewer = "Reviewer"
)

func IsAdmin(role string) bool    { return role == RoleAdmin }
func IsEditor(role string) bool   { return role == RoleEditor }
func IsReviewer(role string) bool { return role == RoleReviewer }
