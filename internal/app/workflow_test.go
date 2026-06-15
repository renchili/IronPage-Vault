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
        {"", ""},
    }
    for _, tc := range cases {
        if got := nextWorkflowStatus(tc.current); got != tc.want {
            t.Fatalf("nextWorkflowStatus(%q)=%q want %q", tc.current, got, tc.want)
        }
    }
}

func TestWorkflowChainIsLinear(t *testing.T) {
    seen := map[string]bool{}
    cur := StatusDraft
    for cur != "" {
        if seen[cur] { t.Fatalf("workflow chain looped at %q", cur) }
        seen[cur] = true
        cur = nextWorkflowStatus(cur)
    }
    for _, required := range []string{StatusDraft, StatusUnderReview, StatusRedactionPending, StatusApproved, StatusFinalized} {
        if !seen[required] { t.Fatalf("workflow chain missing %q", required) }
    }
}
