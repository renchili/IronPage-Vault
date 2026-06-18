package service

import (
	"testing"

	"ironpage-vault/internal/platform"
)

func TestClassifyTextBlockChanges(t *testing.T) {
	unchangedPosition := platform.TextBlock{Page: 1, XMin: 10, YMin: 10, XMax: 20, YMax: 20}
	movedFrom := platform.TextBlock{Page: 1, XMin: 30, YMin: 10, XMax: 40, YMax: 20, Text: "moved"}
	movedTo := platform.TextBlock{Page: 1, XMin: 50, YMin: 10, XMax: 60, YMax: 20, Text: "moved"}

	added := []platform.TextBlock{
		{Page: unchangedPosition.Page, XMin: unchangedPosition.XMin, YMin: unchangedPosition.YMin, XMax: unchangedPosition.XMax, YMax: unchangedPosition.YMax, Text: "after"},
		movedTo,
	}
	removed := []platform.TextBlock{
		{Page: unchangedPosition.Page, XMin: unchangedPosition.XMin, YMin: unchangedPosition.YMin, XMax: unchangedPosition.XMax, YMax: unchangedPosition.YMax, Text: "before"},
		movedFrom,
	}

	remainingAdded, remainingRemoved, modified := classifyTextBlockChanges(added, removed, nil)
	if len(modified) != 1 || modified[0].Text != "after" {
		t.Fatalf("modified=%#v", modified)
	}
	if len(remainingAdded) != 1 || remainingAdded[0].Text != "moved" || remainingAdded[0].XMin != movedTo.XMin {
		t.Fatalf("added=%#v", remainingAdded)
	}
	if len(remainingRemoved) != 1 || remainingRemoved[0].Text != "moved" || remainingRemoved[0].XMin != movedFrom.XMin {
		t.Fatalf("removed=%#v", remainingRemoved)
	}
}

func TestClassifyTextBlockChangesPreservesExistingModified(t *testing.T) {
	existing := platform.TextBlock{Page: 2, XMin: 1, YMin: 2, XMax: 3, YMax: 4, Text: "existing"}
	_, _, modified := classifyTextBlockChanges(nil, nil, []platform.TextBlock{existing})
	if len(modified) != 1 || modified[0] != existing {
		t.Fatalf("modified=%#v", modified)
	}
}
