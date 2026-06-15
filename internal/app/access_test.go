package app

import "testing"

func TestCanReadDocumentObject(t *testing.T) {
    draft := Document{ID:"d1", OwnerID:"editor1", Status:StatusDraft}
    review := Document{ID:"d2", OwnerID:"editor1", Status:StatusUnderReview}
    if !canReadDocumentObject(Principal{Role:RoleAdmin}, draft) { t.Fatalf("admin should read draft") }
    if !canReadDocumentObject(Principal{Role:RoleEditor, UserID:"editor1"}, draft) { t.Fatalf("owner editor should read draft") }
    if canReadDocumentObject(Principal{Role:RoleReviewer, UserID:"rev1"}, draft) { t.Fatalf("reviewer should not read draft") }
    if !canReadDocumentObject(Principal{Role:RoleReviewer, UserID:"rev1"}, review) { t.Fatalf("reviewer should read non-draft") }
}

func TestCanEditDocumentObject(t *testing.T) {
    d := Document{OwnerID:"editor1", Status:StatusDraft}
    if !canEditDocumentObject(Principal{Role:RoleEditor, UserID:"editor1"}, d) { t.Fatalf("owner editor should edit draft") }
    if canEditDocumentObject(Principal{Role:RoleEditor, UserID:"other"}, d) { t.Fatalf("non-owner editor should not edit") }
    if canEditDocumentObject(Principal{Role:RoleAdmin, UserID:"admin"}, d) { t.Fatalf("admin is not editor for document mutation") }
    d.Status = StatusFinalized
    if canEditDocumentObject(Principal{Role:RoleEditor, UserID:"editor1"}, d) { t.Fatalf("finalized document should not be editable") }
}

func TestCanReviewDocumentObject(t *testing.T) {
    if canReviewDocumentObject(Principal{Role:RoleReviewer}, Document{Status:StatusDraft}) { t.Fatalf("draft should not be reviewable") }
    if !canReviewDocumentObject(Principal{Role:RoleReviewer}, Document{Status:StatusUnderReview}) { t.Fatalf("under review should be reviewable") }
    if canReviewDocumentObject(Principal{Role:RoleReviewer}, Document{Status:StatusFinalized}) { t.Fatalf("finalized should not be reviewable") }
}
