package service

import (
	"fmt"

	"ironpage-vault/internal/platform"
)

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
	added, removed, modified = classifyTextBlockChanges(added, removed, modified)
	result["added"] = added
	result["removed"] = removed
	result["modified"] = modified
	return result
}

// classifyTextBlockChanges converts an add/remove pair at the same extracted
// bounding box into one modified block. A block moved to a different box
// remains an add and remove, because its document position changed.
func classifyTextBlockChanges(added, removed, modified []platform.TextBlock) ([]platform.TextBlock, []platform.TextBlock, []platform.TextBlock) {
	removedByPosition := make(map[string]int, len(removed))
	for index, block := range removed {
		removedByPosition[textBlockPositionKey(block)] = index
	}

	usedRemoved := make(map[int]bool)
	remainingAdded := make([]platform.TextBlock, 0, len(added))
	for _, block := range added {
		index, found := removedByPosition[textBlockPositionKey(block)]
		if !found || usedRemoved[index] {
			remainingAdded = append(remainingAdded, block)
			continue
		}
		usedRemoved[index] = true
		modified = append(modified, block)
	}

	remainingRemoved := make([]platform.TextBlock, 0, len(removed))
	for index, block := range removed {
		if !usedRemoved[index] {
			remainingRemoved = append(remainingRemoved, block)
		}
	}
	return remainingAdded, remainingRemoved, modified
}

func textBlockPositionKey(block platform.TextBlock) string {
	return fmt.Sprintf("%d|%.2f|%.2f|%.2f|%.2f", block.Page, block.XMin, block.YMin, block.XMax, block.YMax)
}
