package app

import "testing"

func TestVersionComparisonResultDeclaresBinaryMetadataOnly(t *testing.T) {
	left := DocumentVersion{ID: "left", FileSHA256: "a", SizeBytes: 3, PageCount: 1}
	right := DocumentVersion{ID: "right", FileSHA256: "b", SizeBytes: 4, PageCount: 1}
	result := versionComparisonResult(left, right, []byte("abc"), []byte("abcd"))

	if result["comparison_kind"] != "binary_metadata" {
		t.Fatalf("comparison_kind=%v", result["comparison_kind"])
	}
	if result["text_diff_supported"] != false {
		t.Fatalf("text_diff_supported=%v", result["text_diff_supported"])
	}
	if result["bbox_supported"] != false {
		t.Fatalf("bbox_supported=%v", result["bbox_supported"])
	}
	if result["byte_length_delta"] != 1 {
		t.Fatalf("byte_length_delta=%v", result["byte_length_delta"])
	}
}
