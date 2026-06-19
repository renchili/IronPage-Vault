package app

import "testing"

func TestExtractMentionUsernames(t *testing.T) {
	got := extractMentionUsernames("please review @alice, @bob and @alice again")
	if len(got) != 2 {
		t.Fatalf("expected 2 mentions, got %d: %#v", len(got), got)
	}
	if got[0] != "alice" || got[1] != "bob" {
		t.Fatalf("unexpected mentions: %#v", got)
	}
}

func TestExtractMentionUsernamesIgnoresNonMentions(t *testing.T) {
	got := extractMentionUsernames("email a@b.example and plain words")
	if len(got) != 0 {
		t.Fatalf("unexpected mentions: %#v", got)
	}
}

func TestExtractMentionUsernamesTrimsPunctuation(t *testing.T) {
	got := extractMentionUsernames("(@reviewer-one), please check @legal.team.")
	if len(got) != 2 {
		t.Fatalf("expected 2 mentions, got %#v", got)
	}
	if got[0] != "reviewer-one" || got[1] != "legal.team" {
		t.Fatalf("unexpected mentions: %#v", got)
	}
}
