package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"

	"mddiff/pkg/domain"
)

// Reporter defines the interface for reporting diff results
type Reporter interface {
	Report(report *domain.DiffReport, writer io.Writer) error
}

// NewReporter creates a reporter based on the format string
func NewReporter(format string) (Reporter, error) {
	switch format {
	case "json":
		return &JSONReporter{}, nil
	case "table":
		return &TableReporter{}, nil
	case "markdown":
		return &MarkdownReporter{}, nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}

// JSONReporter outputs the report as JSON
type JSONReporter struct{}

func (r *JSONReporter) Report(report *domain.DiffReport, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// TableReporter outputs the report as a CLI table
type TableReporter struct{}

func (r *TableReporter) Report(report *domain.DiffReport, writer io.Writer) error {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Status", "Path", "Details"})
	table.SetBorder(false) // Set to false for a cleaner look or true if preferred
	table.SetAutoWrapText(false)

	// Sort items by path or status? Spec doesn't strictly say, but usually good practice.
	// For now, iterate as provided.

	for _, item := range report.Items {
		var statusColor []int
		switch item.Type {
		case domain.Missing:
			// Red
			statusColor = []int{tablewriter.Bold, tablewriter.FgRedColor}
		case domain.Extra:
			// Green
			statusColor = []int{tablewriter.Bold, tablewriter.FgGreenColor}
		case domain.Modified:
			// Yellow
			statusColor = []int{tablewriter.Bold, tablewriter.FgYellowColor}
		}

		statusStr := string(item.Type)
		details := item.Reason
		if item.Type == domain.Extra {
			details = fmt.Sprintf("Size: %d bytes", item.TgtSize)
		} else if item.Type == domain.Missing {
			details = fmt.Sprintf("Size: %d bytes", item.SrcSize)
		}

		row := []string{statusStr, item.Path, details}

		// Apply color to the first column (Status)
		table.Rich(row, []tablewriter.Colors{statusColor, {}, {}})
	}

	table.Render()

	// Print Summary below table
	fmt.Fprintf(writer, "\nSummary: Missing: %d, Modified: %d\n", report.Summary.TotalMissing, report.Summary.TotalModified)
	return nil
}

// MarkdownReporter outputs the report as Markdown
type MarkdownReporter struct{}

func (r *MarkdownReporter) Report(report *domain.DiffReport, writer io.Writer) error {
	fmt.Fprintf(writer, "# Diff Report\n\n")
	fmt.Fprintf(writer, "**Source:** `%s`\n", report.SourceDir)
	fmt.Fprintf(writer, "**Target:** `%s`\n\n", report.TargetDir)

	// Group by Type for better Markdown readability
	missing := []domain.DiffItem{}
	extra := []domain.DiffItem{}
	modified := []domain.DiffItem{}

	for _, item := range report.Items {
		switch item.Type {
		case domain.Missing:
			missing = append(missing, item)
		case domain.Extra:
			extra = append(extra, item)
		case domain.Modified:
			modified = append(modified, item)
		}
	}

	if len(missing) > 0 {
		fmt.Fprintf(writer, "## Missing Files (In Source, Not Target)\n")
		for _, item := range missing {
			fmt.Fprintf(writer, "- `%s` (Size: %d)\n", item.Path, item.SrcSize)
		}
		fmt.Fprintln(writer)
	}

	if len(modified) > 0 {
		fmt.Fprintf(writer, "## Modified Files\n")
		for _, item := range modified {
			fmt.Fprintf(writer, "- `%s`: %s\n", item.Path, item.Reason)
		}
		fmt.Fprintln(writer)
	}

	if len(extra) > 0 {
		fmt.Fprintf(writer, "## Extra Files (In Target, Not Source)\n")
		for _, item := range extra {
			fmt.Fprintf(writer, "- `%s` (Size: %d)\n", item.Path, item.TgtSize)
		}
		fmt.Fprintln(writer)
	}

	return nil
}
