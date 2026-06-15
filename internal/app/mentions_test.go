package app

import "testing"

func TestExtractMentionUsernames(t *testing.T) {
    got := extractMentionUsernames("please review @alice, @bob and @alice again")
    if len(got) != 2 { t.Fatalf("expected 2 mentions, got %d: %#v", len(got), got) }
    if got[0] != "alice" || got[1] != "bob" { t.Fatalf("unexpected mentions: %#v", got) }
}

func TestExtractMentionUsernamesIgnoresInvalidTokens(t *testing.T) {
    got := extractMentionUsernames("email a@b.com and @ and @-ok")
    if len(got) != 1 || got[0] != "-ok" { t.Fatalf("unexpected mentions: %#v", got) }
}
