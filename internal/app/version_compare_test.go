package app

import "testing"

func TestVersionComparisonResultIdentical(t *testing.T) {
	v := DocumentVersion{ID: "v1", FileSHA256: "same", SizeBytes: 3, PageCount: 1}
	result := versionComparisonResult(v, v, []byte("abc"), []byte("abc"))
	if result["same_sha256"] != true {
		t.Fatalf("expected same sha")
	}
	if result["same_size"] != true {
		t.Fatalf("expected same size")
	}
	if result["same_page_count"] != true {
		t.Fatalf("expected same page count")
	}
	if result["byte_length_delta"] != 0 {
		t.Fatalf("expected zero delta")
	}
}

func TestVersionComparisonResultDifferentBytes(t *testing.T) {
	left := DocumentVersion{ID: "left", FileSHA256: "a", SizeBytes: 3, PageCount: 1}
	right := DocumentVersion{ID: "right", FileSHA256: "b", SizeBytes: 4, PageCount: 1}
	result := versionComparisonResult(left, right, []byte("abc"), []byte("abcd"))
	if result["same_sha256"] != false {
		t.Fatalf("expected different sha")
	}
	if result["same_size"] != false {
		t.Fatalf("expected different size")
	}
	if result["byte_length_delta"] != 1 {
		t.Fatalf("expected delta 1")
	}
	modified, ok := result["modified"].([]map[string]interface{})
	if !ok || len(modified) != 1 {
		t.Fatalf("expected one modified segment, got %#v", result["modified"])
	}
}
