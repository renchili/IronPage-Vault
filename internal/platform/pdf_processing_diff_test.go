package platform

import "testing"

func TestDiffTextBlocksClassifiesContentChanges(t *testing.T) {
	left := []TextBlock{
		{Page: 1, XMin: 10, YMin: 10, XMax: 20, YMax: 20, Text: "unchanged"},
		{Page: 1, XMin: 30, YMin: 10, XMax: 40, YMax: 20, Text: "before"},
		{Page: 1, XMin: 50, YMin: 10, XMax: 60, YMax: 20, Text: "removed"},
	}
	right := []TextBlock{
		{Page: 1, XMin: 10, YMin: 10, XMax: 20, YMax: 20, Text: "unchanged"},
		{Page: 1, XMin: 30, YMin: 10, XMax: 40, YMax: 20, Text: "after"},
		{Page: 1, XMin: 70, YMin: 10, XMax: 80, YMax: 20, Text: "added"},
	}

	added, removed, modified := DiffTextBlocks(left, right)
	if len(added) != 1 || added[0].Text != "added" {
		t.Fatalf("added=%#v", added)
	}
	if len(removed) != 1 || removed[0].Text != "removed" {
		t.Fatalf("removed=%#v", removed)
	}
	if len(modified) != 1 || modified[0].Text != "after" {
		t.Fatalf("modified=%#v", modified)
	}
}

func TestDiffTextBlocksKeepsChangedPositionAsAddAndRemove(t *testing.T) {
	left := []TextBlock{{Page: 1, XMin: 10, YMin: 10, XMax: 20, YMax: 20, Text: "moved"}}
	right := []TextBlock{{Page: 1, XMin: 30, YMin: 10, XMax: 40, YMax: 20, Text: "moved"}}

	added, removed, modified := DiffTextBlocks(left, right)
	if len(added) != 1 || len(removed) != 1 || len(modified) != 0 {
		t.Fatalf("added=%#v removed=%#v modified=%#v", added, removed, modified)
	}
}
