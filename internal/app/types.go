package app

import "time"

const (
    RoleAdmin = "Admin"
    RoleEditor = "Editor"
    RoleReviewer = "Reviewer"

    StatusDraft = "Draft"
    StatusUnderReview = "Under Review"
    StatusRedactionPending = "Redaction Pending"
    StatusApproved = "Approved"
    StatusFinalized = "Finalized"
)

type User struct {
    ID string `db:"id" json:"id"`
    Username string `db:"username" json:"username"`
    DisplayName string `db:"display_name" json:"display_name"`
    Role string `db:"role" json:"role"`
    PasswordHash string `db:"password_hash" json:"-"`
    FailedAttempts int `db:"failed_attempts" json:"-"`
    LockedUntil *time.Time `db:"locked_until" json:"-"`
}

type Document struct {
    ID string `db:"id" json:"id"`
    Title string `db:"title" json:"title"`
    Status string `db:"status" json:"status"`
    OwnerID string `db:"owner_id" json:"owner_id"`
    CurrentVersion int `db:"current_version" json:"current_version"`
    FinalizedAt *time.Time `db:"finalized_at" json:"finalized_at,omitempty"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type DocumentVersion struct {
    ID string `db:"id" json:"id"`
    DocumentID string `db:"document_id" json:"document_id"`
    VersionNumber int `db:"version_number" json:"version_number"`
    FilePath string `db:"file_path" json:"file_path"`
    FileSHA256 string `db:"file_sha256" json:"file_sha256"`
    SizeBytes int64 `db:"size_bytes" json:"size_bytes"`
    PageCount int `db:"page_count" json:"page_count"`
    CreatedBy string `db:"created_by" json:"created_by"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Principal struct {
    UserID string
    Username string
    Role string
    JTI string
}
