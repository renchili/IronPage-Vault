package app

import "testing"

func TestNormalizeWorkflowDefinitionsAssignsOrderedPositions(t *testing.T) {
	definitions, err := normalizeWorkflowDefinitions([]workflowDefinitionInput{
		{Name: StatusDraft, Mutable: true},
		{Name: StatusUnderReview, Mutable: true},
		{Name: "Quality Review", Mutable: true},
		{Name: StatusFinalized, Mutable: false},
	})
	if err != nil {
		t.Fatalf("normalize workflow: %v", err)
	}
	for index, definition := range definitions {
		if definition.Position != index+1 {
			t.Fatalf("position %d = %d", index, definition.Position)
		}
	}
	if definitions[2].Name != "Quality Review" {
		t.Fatalf("custom status was not retained")
	}
}

func TestNormalizeWorkflowDefinitionsRejectsMutableFinalized(t *testing.T) {
	_, err := normalizeWorkflowDefinitions([]workflowDefinitionInput{
		{Name: StatusDraft, Mutable: true},
		{Name: StatusFinalized, Mutable: true},
	})
	if err == nil {
		t.Fatalf("mutable Finalized must be rejected")
	}
}

func TestNormalizeWorkflowDefinitionsRejectsDuplicateNames(t *testing.T) {
	_, err := normalizeWorkflowDefinitions([]workflowDefinitionInput{
		{Name: StatusDraft, Mutable: true},
		{Name: "Review", Mutable: true},
		{Name: " review ", Mutable: true},
		{Name: StatusFinalized, Mutable: false},
	})
	if err == nil {
		t.Fatalf("case-insensitive duplicate names must be rejected")
	}
}

func TestNormalizeWorkflowDefinitionsRequiresTerminalBoundaries(t *testing.T) {
	cases := [][]workflowDefinitionInput{
		{{Name: "New", Mutable: true}, {Name: StatusFinalized, Mutable: false}},
		{{Name: StatusDraft, Mutable: true}, {Name: "Done", Mutable: false}},
	}
	for _, inputs := range cases {
		if _, err := normalizeWorkflowDefinitions(inputs); err == nil {
			t.Fatalf("invalid workflow boundaries must be rejected: %#v", inputs)
		}
	}
}
