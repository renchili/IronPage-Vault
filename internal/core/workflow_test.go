package core

import "testing"

func TestNextWorkflowStatus(t *testing.T) {
	cases := []struct {
		current string
		want    string
	}{
		{StatusDraft, StatusUnderReview},
		{StatusUnderReview, StatusRedactionPending},
		{StatusRedactionPending, StatusApproved},
		{StatusApproved, StatusFinalized},
		{StatusFinalized, ""},
		{"Unknown", ""},
		{"", ""},
	}

	for _, tc := range cases {
		if got := NextWorkflowStatus(tc.current); got != tc.want {
			t.Fatalf("NextWorkflowStatus(%q)=%q want %q", tc.current, got, tc.want)
		}
	}
}

func TestWorkflowStatusChain(t *testing.T) {
	got := WorkflowStatusChain()
	want := []string{StatusDraft, StatusUnderReview, StatusRedactionPending, StatusApproved, StatusFinalized}

	if len(got) != len(want) {
		t.Fatalf("chain length=%d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("chain[%d]=%q want %q", i, got[i], want[i])
		}
	}
}
