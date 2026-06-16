package service

import "ironpage-vault/internal/platform"

type VersionFile struct {
	ID        string
	FilePath  string
	SHA256    string
	SizeBytes int64
	PageCount int
}

func CompareVersionFiles(left VersionFile, right VersionFile) map[string]interface{} {
	result := map[string]interface{}{
		"left_version_id":     left.ID,
		"right_version_id":    right.ID,
		"comparison_kind":     "text_bbox",
		"text_diff_supported": true,
		"bbox_supported":      true,
		"same_sha256":         left.SHA256 == right.SHA256,
		"same_size":           left.SizeBytes == right.SizeBytes,
		"same_page_count":     left.PageCount == right.PageCount,
		"byte_length_delta":   right.SizeBytes - left.SizeBytes,
		"added":               []platform.TextBlock{},
		"removed":             []platform.TextBlock{},
		"modified":            []platform.TextBlock{},
	}
	leftBlocks, leftMode, leftErr := platform.ExtractPDFTextBlocks(left.FilePath)
	rightBlocks, rightMode, rightErr := platform.ExtractPDFTextBlocks(right.FilePath)
	result["text_extract_left_mode"] = leftMode
	result["text_extract_right_mode"] = rightMode
	if leftErr != nil || rightErr != nil {
		result["text_diff_supported"] = false
		result["bbox_supported"] = false
		return result
	}
	added, removed, modified := platform.DiffTextBlocks(leftBlocks, rightBlocks)
	result["added"] = added
	result["removed"] = removed
	result["modified"] = modified
	return result
}
