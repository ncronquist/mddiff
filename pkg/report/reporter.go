// Package report handles formatting the results of the diff.
package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"

	"mddiff/pkg/domain"
)

// Reporter defines the interface for reporting diff results.
type Reporter interface {
	Report(report *domain.DiffReport, writer io.Writer) error
}

// NewReporter creates a reporter based on the format string.
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

// JSONReporter outputs the report as JSON.
type JSONReporter struct{}

// Report implements the Reporter interface for JSON output.
func (r *JSONReporter) Report(report *domain.DiffReport, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// TableReporter outputs the report as a CLI table.
type TableReporter struct{}

// Report implements the Reporter interface for Table output.
func (r *TableReporter) Report(report *domain.DiffReport, writer io.Writer) error {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Status", "Path", "Details"})
	// tablewriter v0.0.5 doesn't assume default border styles same as new ones,
	// checking docs or source previously showed SetBorder exists but let's stick to basic defaults if we had issues.
	// Actually v0.0.5 has SetBorder. The issue before was likely implicit v1.x vs v0.x confusion.
	// Let's keep it simple.
	table.SetBorder(false)
	table.SetAutoWrapText(false)

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
		switch item.Type {
		case domain.Extra:
			details = fmt.Sprintf("Size: %d bytes", item.TgtSize)
		case domain.Missing:
			details = fmt.Sprintf("Size: %d bytes", item.SrcSize)
		}

		row := []string{statusStr, item.Path, details}

		// Apply color to the first column (Status)
		table.Rich(row, []tablewriter.Colors{statusColor, {}, {}})
	}

	table.Render()

	// Print Summary below table
	// We must check errors for errcheck linter
	if _, err := fmt.Fprintf(
		writer,
		"\nSummary: Missing: %d, Modified: %d\n",
		report.Summary.TotalMissing,
		report.Summary.TotalModified,
	); err != nil {
		return err
	}
	return nil
}

// MarkdownReporter outputs the report as Markdown.
type MarkdownReporter struct{}

// Report implements the Reporter interface for Markdown output.
func (r *MarkdownReporter) Report(report *domain.DiffReport, writer io.Writer) error {
	if _, err := fmt.Fprintf(writer, "# Diff Report\n\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "**Source:** `%s`\n", report.SourceDir); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "**Target:** `%s`\n\n", report.TargetDir); err != nil {
		return err
	}

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

	if err := r.printMissing(writer, missing); err != nil {
		return err
	}

	if err := r.printModified(writer, modified); err != nil {
		return err
	}

	if err := r.printExtra(writer, extra); err != nil {
		return err
	}

	return nil
}

func (r *MarkdownReporter) printMissing(writer io.Writer, items []domain.DiffItem) error {
	if len(items) == 0 {
		return nil
	}
	if _, err := fmt.Fprintf(writer, "## Missing Files (In Source, Not Target)\n"); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := fmt.Fprintf(writer, "- `%s` (Size: %d)\n", item.Path, item.SrcSize); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	return nil
}

func (r *MarkdownReporter) printModified(writer io.Writer, items []domain.DiffItem) error {
	if len(items) == 0 {
		return nil
	}
	if _, err := fmt.Fprintf(writer, "## Modified Files\n"); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := fmt.Fprintf(writer, "- `%s`: %s\n", item.Path, item.Reason); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	return nil
}

func (r *MarkdownReporter) printExtra(writer io.Writer, items []domain.DiffItem) error {
	if len(items) == 0 {
		return nil
	}
	if _, err := fmt.Fprintf(writer, "## Extra Files (In Target, Not Source)\n"); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := fmt.Fprintf(writer, "- `%s` (Size: %d)\n", item.Path, item.TgtSize); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	return nil
}
