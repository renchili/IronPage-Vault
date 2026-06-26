package core

// NextWorkflowStatus returns the next status in the mandatory document lifecycle.
//
// It returns an empty string when current is unknown or already terminal. Keeping
// this rule in core ensures handlers and services share the same transition order.
func NextWorkflowStatus(current string) string {
	chain := WorkflowStatusChain()
	for i, s := range chain {
		if s == current && i+1 < len(chain) {
			return chain[i+1]
		}
	}
	return ""
}

// WorkflowStatusChain returns the canonical legal-document lifecycle order.
func WorkflowStatusChain() []string {
	return []string{StatusDraft, StatusUnderReview, StatusRedactionPending, StatusApproved, StatusFinalized}
}
