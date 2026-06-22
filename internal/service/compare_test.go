package service

import (
	"os"
	"path/filepath"
	"testing"

	"ironpage-vault/internal/platform"
)

func TestCompareVersionFilesReportsAddedRemovedAndModifiedText(t *testing.T) {
	dir := t.TempDir()
	leftText := "UNCHANGED\nOLD LINE\nREMOVED ONLY\n"
	rightText := "UNCHANGED\nNEW LINE\n\nADDED ONLY\n"
	leftPath := filepath.Join(dir, "left.txt")
	rightPath := filepath.Join(dir, "right.txt")
	if err := os.WriteFile(leftPath, []byte(leftText), 0600); err != nil {
		t.Fatalf("write left: %v", err)
	}
	if err := os.WriteFile(rightPath, []byte(rightText), 0600); err != nil {
		t.Fatalf("write right: %v", err)
	}

	result := CompareVersionFiles(
		VersionFile{ID: "ver_left", FilePath: leftPath, SHA256: "left-sha", SizeBytes: int64(len(leftText)), PageCount: 1},
		VersionFile{ID: "ver_right", FilePath: rightPath, SHA256: "right-sha", SizeBytes: int64(len(rightText)), PageCount: 1},
	)

	if got := result["comparison_kind"]; got != "text_bbox" {
		t.Fatalf("comparison_kind = %v, want text_bbox", got)
	}
	if got := result["text_diff_supported"]; got != true {
		t.Fatalf("text_diff_supported = %v, want true", got)
	}
	if got := result["bbox_supported"]; got != true {
		t.Fatalf("bbox_supported = %v, want true", got)
	}
	assertBlocksContainText(t, result["added"], "ADDED ONLY")
	assertBlocksContainText(t, result["removed"], "REMOVED ONLY")
	assertBlocksContainText(t, result["modified"], "NEW LINE")
}

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

func assertBlocksContainText(t *testing.T, raw interface{}, want string) {
	t.Helper()
	blocks, ok := raw.([]platform.TextBlock)
	if !ok {
		t.Fatalf("%T is not []platform.TextBlock", raw)
	}
	for _, block := range blocks {
		if block.Text == want {
			return
		}
	}
	t.Fatalf("blocks %+v do not contain text %q", blocks, want)
}
