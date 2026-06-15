package app

import "testing"

func TestNextWorkflowStatus(t *testing.T) {
    cases := []struct{
        current string
        want string
    }{
        {StatusDraft, StatusUnderReview},
        {StatusUnderReview, StatusRedactionPending},
        {StatusRedactionPending, StatusApproved},
        {StatusApproved, StatusFinalized},
        {StatusFinalized, ""},
        {"Unknown", ""},
    }
    for _, tc := range cases {
        if got := nextWorkflowStatus(tc.current); got != tc.want {
            t.Fatalf("nextWorkflowStatus(%q)=%q want %q", tc.current, got, tc.want)
        }
    }
}
