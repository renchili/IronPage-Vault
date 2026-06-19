package store

import "testing"

func TestDocumentListWhereClause(t *testing.T) {
	adminWhere, adminArgs := DocumentListWhereClause(PrincipalFilter{Role: RoleAdmin})
	if adminWhere != "1=1" || adminArgs != nil {
		t.Fatalf("admin filter mismatch: %q %#v", adminWhere, adminArgs)
	}

	reviewerWhere, reviewerArgs := DocumentListWhereClause(PrincipalFilter{Role: RoleReviewer})
	if reviewerWhere != "status <> 'Draft'" || reviewerArgs != nil {
		t.Fatalf("reviewer filter mismatch: %q %#v", reviewerWhere, reviewerArgs)
	}

	editorWhere, editorArgs := DocumentListWhereClause(PrincipalFilter{Role: "Editor", UserID: "u1"})
	if editorWhere != "owner_id=$1" {
		t.Fatalf("editor where=%q", editorWhere)
	}
	if len(editorArgs) != 1 || editorArgs[0] != "u1" {
		t.Fatalf("editor args=%#v", editorArgs)
	}
}
