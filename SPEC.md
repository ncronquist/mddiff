# Project Specification: mddiff (Media Directory Diff)

## 1\. Project Overview

`mddiff` is a CLI tool written in Go to generate a differential report between two media directories (Source vs. Target). It identifies missing directories, missing files, and modified media.

**Key Philosophy:**

1.  **Strict Naming:** File identity is tied strictly to the filename stem. `Movie (2000).mkv` and `Movie.mp4` are considered two distinct files (one removed, one added), not a modification.
2.  **Extensible Design:** While V1 uses a linear scan and simple file-stat comparison, the code must be structured via interfaces to support concurrent scanning and deep `ffprobe` inspection in V2 without refactoring the core logic.

## 2\. Architectural Guidelines

The project uses a standard Go layout with a strict separation of CLI and Business Logic.

### 2.1 Directory Structure

```text
├── cmd/            # CLI wiring (Cobra) - No business logic.
├── pkg/
│   ├── domain/     # Shared Structs, Interfaces, and Constants.
│   ├── scanner/    # logic to walk the filesystem.
│   ├── diff/       # Logic to compare two directory trees.
│   └── report/     # Logic to render output (JSON, Table, etc).
└── main.go         # Entrypoint.
```

### 2.2 Interface-Driven Design (Crucial for V2)

To satisfy the requirement of future upgrades (concurrency and deep inspection), the core logic must rely on interfaces defined in `pkg/domain`.

**The Scanner Interface:**
Allows swapping the linear walker for a concurrent one later.

```go
type Scanner interface {
    Scan(rootPath string) (*domain.DirectoryTree, error)
}
```

**The Comparator Interface:**
Allows swapping simple `Stat` checks for deep `ffprobe` checks later.

```go
type AssetComparator interface {
    // Returns true if assets are considered "content modified"
    // (e.g. size changed, or codec changed in V2)
    Compare(src, tgt domain.Asset) (bool, string)
}
```

## 3\. Functional Requirements

### 3.1 The Scanner (`pkg/scanner`)

  * **Walk Strategy:** Linear walk (recursive or `filepath.WalkDir`) for V1.
  * **Ignore List:**
      * The scanner **must** skip files/directories based on a hardcoded list for V1.
      * **V1 Ignore List:** `.DS_Store`, `Thumbs.db`, `.git`, `.idea`, `.vscode`.
      * *Implementation Note:* Define this list as a variable (slice of strings) that can be easily exposed to a config file loader in the future.
  * **Data Collection:**
      * Relative Path (key for comparison).
      * Filename Stem (filename without extension).
      * Extension.
      * Size (bytes).

### 3.2 The Diff Engine (`pkg/diff`)

The engine compares the Source tree against the Target tree.

**Matching Rules:**

1.  **Identity:** An asset is identified by its **Relative Path** + **Filename Stem**.
      * `src/A/movie.mkv` matches `tgt/A/movie.mp4` (Same stem "movie").
      * `src/A/movie (2000).mkv` DOES NOT match `tgt/A/movie.mp4` (Stems differ).
2.  **Comparison Logic:**
      * **MISSING:** Stem exists in Source, but not in Target.
      * **EXTRA:** Stem exists in Target, but not in Source.
      * **MODIFIED:** Stem exists in both, but attributes differ.
          * *Attribute 1:* Extension changed (e.g., `.mkv` -\> `.mp4`).
          * *Attribute 2:* Size changed (using a configurable threshold, e.g., \>100 bytes difference).

### 3.3 The Reporter (`pkg/report`)

  * **Output:** Stdout.
  * **Formats:**
      * `json`: Complete dump of the diff struct.
      * `table`: Columns -\> [Status, Path, Details].
      * `markdown`: Grouped by status (e.g., "\#\#\# Missing Files").

## 4\. Data Structures (`pkg/domain`)

```go
package domain

// Asset represents a single file
type Asset struct {
    Path      string // Relative path including filename
    Stem      string // Filename without extension
    Extension string // .mkv, .mp4, .srt
    Size      int64
    IsDir     bool
}

// DiffType enums
type DiffType string
const (
    Missing  DiffType = "MISSING"  // In Source, not Target
    Extra    DiffType = "EXTRA"    // In Target, not Source
    Modified DiffType = "MODIFIED" // In both, content/ext differs
)

type DiffItem struct {
    Type    DiffType `json:"type"`
    Path    string   `json:"path"`
    Reason  string   `json:"reason,omitempty"` // e.g. "Size changed", "mkv -> mp4"
    SrcSize int64    `json:"src_size,omitempty"`
    TgtSize int64    `json:"tgt_size,omitempty"`
}

type DiffReport struct {
    SourceDir string     `json:"source_dir"`
    TargetDir string     `json:"target_dir"`
    Items     []DiffItem `json:"items"`
    Summary   struct {
        TotalMissing  int `json:"total_missing"`
        TotalModified int `json:"total_modified"`
    } `json:"summary"`
}
```

## 5\. CLI Usage & wiring (`cmd/`)

  * **Command:** `mddiff [source] [target]`
  * **Flags:**
      * `--format, -f`: "table" (default), "json", "markdown"
      * `--ignore-ext`: (Optional V1 feature) Comma-separated list of extensions to ignore (e.g., `.nfo,.txt`).
  * **Wiring:**
    1.  Parse flags.
    2.  Instantiate `Scanner` (V1 implementation).
    3.  Run `Scanner.Scan(source)` and `Scanner.Scan(target)`.
    4.  Instantiate `Differ` with a specific `Comparator` (V1 Size/Ext comparator).
    5.  Run `Differ.Diff(srcTree, tgtTree)`.
    6.  Pass result to `Reporter`.

## 6\. Test Plan

All logic in `pkg` must be tested.

1.  **Scanner Tests:**
      * Create a temporary directory structure.
      * Ensure "Ignored Files" (.DS\_Store) are actually skipped.
      * Ensure Stems are calculated correctly.
2.  **Diff Tests:**
      * **Strict Name Check:** Test that `Movie (2022).mkv` vs `Movie.mkv` results in 1 Missing, 1 Extra.
      * **Format Change:** Test that `Video.mkv` vs `Video.mp4` results in 1 Modified.
      * **Size Change:** Test that `Video.mkv` (1GB) vs `Video.mkv` (500MB) results in 1 Modified.