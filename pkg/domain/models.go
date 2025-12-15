// Package domain contains the core data structures and interfaces for mddiff.
package domain

// Asset represents a single file metadata.
type Asset struct {
	Path      string // Relative path including filename
	Stem      string // Filename without extension
	Extension string // .mkv, .mp4, .srt
	Size      int64
	IsDir     bool
}

// DirectoryTree represents the result of a scan.
type DirectoryTree struct {
	RootPath string
	Assets   map[string]Asset // Keyed by relative path
}

// DiffType enums.
type DiffType string

const (
	// Missing represents a file in Source but not Target.
	Missing DiffType = "MISSING"
	// Extra represents a file in Target but not Source.
	Extra DiffType = "EXTRA"
	// Modified represents a file in both but with different content.
	Modified DiffType = "MODIFIED"
)

// DiffItem represents a single difference found.
type DiffItem struct {
	Type    DiffType `json:"type"`
	Path    string   `json:"path"`
	Reason  string   `json:"reason,omitempty"` // e.g. "Size changed", "mkv -> mp4"
	SrcSize int64    `json:"src_size,omitempty"`
	TgtSize int64    `json:"tgt_size,omitempty"`
}

// DiffReport represents the full comparison report.
type DiffReport struct {
	SourceDir string     `json:"source_dir"`
	TargetDir string     `json:"target_dir"`
	Items     []DiffItem `json:"items"`
	Summary   struct {
		TotalMissing  int `json:"total_missing"`
		TotalModified int `json:"total_modified"`
	} `json:"summary"`
}

// Scanner defines the interface for walking a directory.
type Scanner interface {
	Scan(rootPath string) (*DirectoryTree, error)
}

// AssetComparator defines the interface for comparing two assets.
type AssetComparator interface {
	// Compare returns true if assets are considered "content modified"
	// (e.g. size changed, or codec changed in V2)
	// Returns modified bool and a reason string
	Compare(src, tgt Asset) (bool, string)
}
