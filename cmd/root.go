/*
Package cmd is the CLI package for mddiff.
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var format string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "mddiff path/to/dir1 path/to/dir2",
	Short: "A media directory diff tool",
	Long: `
mddiff is a tool for comparing two media directories and identifying
differences between them. Since mddiff knows the directories being compared
are media directories, it be smarter about how it compares the files within
them. For example, it can understand the difference between two completely
different video files and two different encodings of the same video file.`,
	Args:    cobra.ExactArgs(2),
	PreRunE: validateInputs,
	Run:     mddiff,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&format, "format", "f", "human", "Output format (human|json)")
}

func mddiff(cmd *cobra.Command, args []string) {
	dir1 := args[0]
	dir2 := args[1]

	format, _ := cmd.Flags().GetString("format")

	dir1abs, err := filepath.Abs(filepath.Clean(dir1))
	if err != nil {
		println("Error getting absolute path for dir1:", err.Error())
		os.Exit(1)
	}
	dir2abs, err := filepath.Abs(filepath.Clean(dir2))
	if err != nil {
		println("Error getting absolute path for dir2:", err.Error())
		os.Exit(1)
	}

	println("Resolved dir1:", dir1abs)
	println("Resolved dir2:", dir2abs)
	println("format:", format)
	// compareDirectories(dir1, dir2, format)
}

func validateInputs(_ *cobra.Command, _ []string) error {
	// Enum check
	switch format {
	case "human", "json":
	default:
		return fmt.Errorf("invalid --format: %s (want human|json)", format)
	}
	return nil
}
