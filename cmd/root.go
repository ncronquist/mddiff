package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mddiff/pkg/diff"
	"mddiff/pkg/report"
	"mddiff/pkg/scanner"
)

var (
	format    string
	ignoreExt string // Comma-separated list for V1 optional feature
	verbose   bool
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "mddiff [source] [target]",
	Short: "A media directory diff tool",
	Long: `mddiff is a tool for comparing two media directories and identifying
differences between them (Missing, Extra, Modified).`,
	Args: cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate Format
		validFormats := map[string]bool{"json": true, "table": true, "markdown": true}
		if !validFormats[format] {
			return fmt.Errorf("invalid format '%s'. Must be one of: json, table, markdown", format)
		}

		// Validate Paths
		if err := validateDir(args[0]); err != nil {
			return fmt.Errorf("source argument error: %w", err)
		}
		if err := validateDir(args[1]); err != nil {
			return fmt.Errorf("target argument error: %w", err)
		}
		return nil
	},
	RunE: runDiff,
}

func validateDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", path)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}
	return nil
}

func runDiff(cmd *cobra.Command, args []string) error {
	sourcePath := args[0]
	targetPath := args[1]

	// 1. Setup Scanner
	// Note: In V1 we use the hardcoded scanner, ignore-ext flag implementation
	// would require passing it to the scanner.
	// For now, we'll stick to the hardcoded list in NewLinearScanner as per "V1 Ignore List" spec.
	// Optionally extending it if ignoreExt is provided is a nice-to-have but staying strict to plan.
	scan := scanner.NewLinearScanner()

	if verbose {
		fmt.Printf("Scanning Source: %s\n", sourcePath)
	}
	sourceTree, err := scan.Scan(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to scan source: %w", err)
	}

	if verbose {
		fmt.Printf("Scanning Target: %s\n", targetPath)
	}
	targetTree, err := scan.Scan(targetPath)
	if err != nil {
		return fmt.Errorf("failed to scan target: %w", err)
	}

	// 2. Setup Diff Engine
	// Basic comparator (Size matching)
	// Threshold could be a flag, but using default 0 or small buffer
	comparator := &diff.BasicComparator{SizeThreshold: 0}
	engine := diff.NewEngine(comparator)

	if verbose {
		fmt.Println("Calculating Diff...")
	}
	diffReport := engine.Diff(sourceTree, targetTree)

	// 3. Report
	reporter, err := report.NewReporter(format)
	if err != nil {
		return err
	}

	return reporter.Report(diffReport, os.Stdout)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table|json|markdown)")
	rootCmd.Flags().StringVar(&ignoreExt, "ignore-ext", "", "Comma-separated list of extensions to ignore (e.g. .txt,.nfo)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}
