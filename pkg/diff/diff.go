package diff

import (
	"path/filepath"

	"mddiff/pkg/domain"
)

// Engine handles the comparison logic
type Engine struct {
	comparator domain.AssetComparator
}

// NewEngine creates a new diff engine with a specific comparator
func NewEngine(comparator domain.AssetComparator) *Engine {
	return &Engine{
		comparator: comparator,
	}
}

// Diff compares two directory trees and produces a report
func (e *Engine) Diff(source, target *domain.DirectoryTree) *domain.DiffReport {
	report := &domain.DiffReport{
		SourceDir: source.RootPath,
		TargetDir: target.RootPath,
		Items:     []domain.DiffItem{},
	}

	// Index target assets by "Stem Identity" for quick lookup
	// Identity = Relative Directory + Filename Stem
	targetIndex := make(map[string]domain.Asset)
	for relPath, asset := range target.Assets {
		id := makeIdentity(relPath, asset.Stem)
		targetIndex[id] = asset
	}

	// Track processed target assets to find "EXTRA" files later
	processedTargets := make(map[string]bool)

	// 1. Iterate through Source Assets to find MISSING and MODIFIED
	for relPath, srcAsset := range source.Assets {
		id := makeIdentity(relPath, srcAsset.Stem)

		tgtAsset, exists := targetIndex[id]

		if !exists {
			// Case: MISSING
			report.Items = append(report.Items, domain.DiffItem{
				Type:    domain.Missing,
				Path:    srcAsset.Path,
				SrcSize: srcAsset.Size,
			})
			report.Summary.TotalMissing++
		} else {
			// Case: EXISTS, check for modifications
			processedTargets[id] = true // Mark as visited

			modified, reason := e.comparator.Compare(srcAsset, tgtAsset)
			if modified {
				report.Items = append(report.Items, domain.DiffItem{
					Type:    domain.Modified,
					Path:    srcAsset.Path,
					Reason:  reason,
					SrcSize: srcAsset.Size,
					TgtSize: tgtAsset.Size,
				})
				report.Summary.TotalModified++
			}
		}
	}

	// 2. Iterate through Target Assets to find EXTRA
	for relPath, tgtAsset := range target.Assets {
		id := makeIdentity(relPath, tgtAsset.Stem)
		if !processedTargets[id] {
			report.Items = append(report.Items, domain.DiffItem{
				Type:    domain.Extra,
				Path:    tgtAsset.Path,
				TgtSize: tgtAsset.Size,
			})
		}
	}

	return report
}

// makeIdentity creates a unique key based on directory and stem
// e.g. "subdir/Movie"
func makeIdentity(relPath, stem string) string {
	dir := filepath.Dir(relPath)
	if dir == "." {
		return stem
	}
	return filepath.Join(dir, stem)
}

// BasicComparator implements AsssetComparator for V1 (Size + Extension)
type BasicComparator struct {
	SizeThreshold int64
}

func (c *BasicComparator) Compare(src, tgt domain.Asset) (bool, string) {
	// Check Extension
	if src.Extension != tgt.Extension {
		return true, "Extension changed: " + src.Extension + " -> " + tgt.Extension
	}

	// Check Size
	diff := src.Size - tgt.Size
	if diff < 0 {
		diff = -diff
	}

	if diff > c.SizeThreshold {
		return true, "Size changed"
	}

	return false, ""
}
