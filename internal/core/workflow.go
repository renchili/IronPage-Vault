package core

func NextWorkflowStatus(current string) string {
	chain := WorkflowStatusChain()
	for i, s := range chain {
		if s == current && i+1 < len(chain) {
			return chain[i+1]
		}
	}
	return ""
}

func WorkflowStatusChain() []string {
	return []string{StatusDraft, StatusUnderReview, StatusRedactionPending, StatusApproved, StatusFinalized}
}
