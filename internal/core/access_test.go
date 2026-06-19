package core

import "testing"

func TestCanReadDocumentObject(t *testing.T) {
	draft := DocumentAccess{OwnerID: "editor1", Status: StatusDraft}
	review := DocumentAccess{OwnerID: "editor1", Status: StatusUnderReview}
	if !CanReadDocumentObject(Principal{Role: RoleAdmin}, draft) {
		t.Fatalf("admin should read draft")
	}
	if !CanReadDocumentObject(Principal{Role: RoleEditor, UserID: "editor1"}, draft) {
		t.Fatalf("owner editor should read draft")
	}
	if CanReadDocumentObject(Principal{Role: RoleEditor, UserID: "other"}, draft) {
		t.Fatalf("non-owner editor should not read draft")
	}
	if CanReadDocumentObject(Principal{Role: RoleReviewer}, draft) {
		t.Fatalf("reviewer should not read draft")
	}
	if !CanReadDocumentObject(Principal{Role: RoleReviewer}, review) {
		t.Fatalf("reviewer should read non-draft")
	}
}

func TestCanEditDocumentObject(t *testing.T) {
	d := DocumentAccess{OwnerID: "editor1", Status: StatusDraft}
	if !CanEditDocumentObject(Principal{Role: RoleEditor, UserID: "editor1"}, d) {
		t.Fatalf("owner editor should edit draft")
	}
	if CanEditDocumentObject(Principal{Role: RoleEditor, UserID: "other"}, d) {
		t.Fatalf("non-owner editor should not edit")
	}
	if CanEditDocumentObject(Principal{Role: RoleAdmin}, d) {
		t.Fatalf("admin should not use editor mutation")
	}
	if CanEditDocumentObject(Principal{Role: RoleReviewer}, d) {
		t.Fatalf("reviewer should not edit")
	}
	d.Status = StatusFinalized
	if CanEditDocumentObject(Principal{Role: RoleEditor, UserID: "editor1"}, d) {
		t.Fatalf("finalized document should not be editable")
	}
}

func TestCanReviewDocumentObject(t *testing.T) {
	if CanReviewDocumentObject(Principal{Role: RoleReviewer}, DocumentAccess{Status: StatusDraft}) {
		t.Fatalf("draft should not be reviewable")
	}
	if !CanReviewDocumentObject(Principal{Role: RoleReviewer}, DocumentAccess{Status: StatusUnderReview}) {
		t.Fatalf("under review should be reviewable")
	}
	if CanReviewDocumentObject(Principal{Role: RoleReviewer}, DocumentAccess{Status: StatusFinalized}) {
		t.Fatalf("finalized should not be reviewable")
	}
	if CanReviewDocumentObject(Principal{Role: RoleEditor}, DocumentAccess{Status: StatusUnderReview}) {
		t.Fatalf("editor should not use reviewer flow")
	}
}

func TestCanTransitionDocumentObject(t *testing.T) {
	ownedDraft := DocumentAccess{OwnerID: "editor1", Status: StatusDraft}
	if !CanTransitionDocumentObject(Principal{Role: RoleEditor, UserID: "editor1"}, ownedDraft) {
		t.Fatalf("owner editor should transition draft")
	}
	if CanTransitionDocumentObject(Principal{Role: RoleEditor, UserID: "other"}, ownedDraft) {
		t.Fatalf("non-owner editor should not transition draft")
	}
	if CanTransitionDocumentObject(Principal{Role: RoleReviewer}, ownedDraft) {
		t.Fatalf("reviewer should not transition draft")
	}
	if !CanTransitionDocumentObject(Principal{Role: RoleReviewer}, DocumentAccess{Status: StatusUnderReview}) {
		t.Fatalf("reviewer should transition reviewable document")
	}
	if CanTransitionDocumentObject(Principal{Role: RoleEditor, UserID: "editor1"}, DocumentAccess{OwnerID: "editor1", Status: StatusFinalized}) {
		t.Fatalf("finalized should not transition")
	}
}
