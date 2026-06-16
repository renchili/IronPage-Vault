package core

import "testing"

func TestExtractMentionUsernames(t *testing.T) {
	got := ExtractMentionUsernames("please review @alice and @bob.")
	want := []string{"alice", "bob"}

	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestExtractMentionUsernamesDeduplicates(t *testing.T) {
	got := ExtractMentionUsernames("@alice @alice @bob")
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestExtractMentionUsernamesIgnoresEmailLikeText(t *testing.T) {
	got := ExtractMentionUsernames("email alice@example.com and @reviewer")
	if len(got) != 1 || got[0] != "reviewer" {
		t.Fatalf("got %v", got)
	}
}

func TestExtractMentionUsernamesTrimsPunctuation(t *testing.T) {
	got := ExtractMentionUsernames("(@legal-team), @reviewer!")
	want := []string{"legal-team", "reviewer"}

	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}
