package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"mddiff/pkg/domain"
)

// LinearScanner implements the Scanner interface using filepath.WalkDir
type LinearScanner struct {
	ignoredFiles map[string]struct{}
}

// NewLinearScanner creates a new scanner with the default ignore list
func NewLinearScanner() *LinearScanner {
	// V1 Ignore List: .DS_Store, Thumbs.db, .git, .idea, .vscode
	ignoreList := []string{".DS_Store", "Thumbs.db", ".git", ".idea", ".vscode"}
	ignored := make(map[string]struct{})
	for _, f := range ignoreList {
		ignored[f] = struct{}{}
	}

	return &LinearScanner{
		ignoredFiles: ignored,
	}
}

// Scan walks the directory and returns a DirectoryTree
func (s *LinearScanner) Scan(rootPath string) (*domain.DirectoryTree, error) {
	assets := make(map[string]domain.Asset)

	// Clean root path to ensure consistent relative paths
	rootPath = filepath.Clean(rootPath)

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		// Skip root directory itself
		if relPath == "." {
			return nil
		}

		// Check ignore list
		if _, ok := s.ignoredFiles[d.Name()]; ok {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		// Determine stem and extension
		ext := filepath.Ext(d.Name())
		stem := strings.TrimSuffix(d.Name(), ext)

		asset := domain.Asset{
			Path:      relPath,
			Stem:      stem,
			Extension: ext,
			Size:      info.Size(),
			IsDir:     d.IsDir(),
		}

		assets[relPath] = asset

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &domain.DirectoryTree{
		RootPath: rootPath,
		Assets:   assets,
	}, nil
}
