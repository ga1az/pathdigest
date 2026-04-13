package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ga1az/pathdigest/internal/digest"
	"github.com/spf13/cobra"
)

var (
	appVersion string
	goVersion  string
)

func SetVersionInfo(appV, goV string) {
	appVersion = appV
	if goV != "" {
		goVersion = goV
	}
}

var (
	outputFile      string
	maxFileSize     int64
	excludePatterns []string
	includePatterns []string
	branch          string
	outputFormat    string
)

var rootCmd = &cobra.Command{
	Use:   "pathdigest <source>",
	Short: "Generates a prompt-friendly text digest of a Git repository or local directory.",
	Long: `pathdigest analyzes a Git repository (from a URL) or a local directory
and creates a structured text output of its codebase.

This output is optimized for use as context for Large Language Models (LLMs).
You can specify a local path or a repository URL as the source.`,
	Args:    cobra.ExactArgs(1),
	Version: appVersion,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		userExcludePatterns, err := cmd.Flags().GetStringSlice("exclude-pattern")
		if err != nil {
			return fmt.Errorf("error parsing exclude patterns: %w", err)
		}

		finalExcludesSet := make(map[string]struct{})

		for _, p := range digest.DefaultExcludePatterns {
			finalExcludesSet[p] = struct{}{}
		}

		for _, p := range userExcludePatterns {
			finalExcludesSet[p] = struct{}{}
		}

		excludePatterns = make([]string, 0, len(finalExcludesSet))
		for p := range finalExcludesSet {
			excludePatterns = append(excludePatterns, p)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Validate format flag early
		if outputFormat != "text" && outputFormat != "json" {
			fmt.Fprintf(os.Stderr, "Error: unsupported format '%s'. Use 'text' or 'json'.\n", outputFormat)
			os.Exit(1)
		}

		source := args[0]

		opts := digest.IngestionOptions{
			Source:          source,
			OutputFile:      outputFile,
			MaxFileSize:     maxFileSize,
			ExcludePatterns: excludePatterns,
			IncludePatterns: includePatterns,
			Branch:          branch,
		}

		fmt.Fprintf(os.Stderr, "Processing source: %s\n", opts.Source)
		if opts.Branch != "" {
			fmt.Fprintf(os.Stderr, "Targeting branch: %s\n", opts.Branch)
		}

		ingestResult, err := digest.ProcessSource(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing source: %v\n", err)
			os.Exit(1)
		}

		// Route to the correct formatter — exactly one call
		if outputFormat == "json" {
			jsonBytes, errJSON := ingestResult.FormatJSON(opts)
			if errJSON != nil {
				fmt.Fprintf(os.Stderr, "Error formatting JSON output: %v\n", errJSON)
				os.Exit(1)
			}

			if opts.OutputFile != "" && opts.OutputFile != "-" {
				if err := writeOutputFile(opts.OutputFile, jsonBytes); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to output file %s: %v\n", opts.OutputFile, err)
					os.Exit(1)
				}
				fmt.Fprintf(os.Stderr, "Digest written to: %s\n", opts.OutputFile)
			} else {
				os.Stdout.Write(jsonBytes)
			}
		} else {
			ingestResult.FormatOutput(opts)

			if opts.OutputFile != "" && opts.OutputFile != "-" {
				textBytes := []byte(ingestResult.TreeStructure + "\n" + ingestResult.FileContents)
				if err := writeOutputFile(opts.OutputFile, textBytes); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to output file %s: %v\n", opts.OutputFile, err)
					os.Exit(1)
				}
				fmt.Fprintf(os.Stderr, "Digest written to: %s\n", opts.OutputFile)
			} else {
				fmt.Println(ingestResult.TreeStructure)
				fmt.Println(ingestResult.FileContents)
			}

			fmt.Fprintln(os.Stderr, "\n--- Summary ---")
			fmt.Fprint(os.Stderr, ingestResult.Summary)
		}

	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of pathdigest",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pathdigest version: %s\n", appVersion)
		if goVersion != "" && goVersion != "unknown" && goVersion != "built with Go" {
			fmt.Printf("Built with Go version: %s\n", goVersion)
		} else if goVersion == "built with Go" {
			fmt.Printf("%s\n", goVersion)
		}
	},
}

func writeOutputFile(path string, data []byte) error {
	outputDir := filepath.Dir(path)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0644)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)

	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "pathdigest_digest.txt", "Output file path")
	rootCmd.Flags().Int64VarP(&maxFileSize, "max-size", "s", 10*1024*1024, "Maximum file size to process in bytes (e.g., 10485760 for 10MB)") // 10MB default

	rootCmd.Flags().StringSliceP("exclude-pattern", "e", []string{}, "Comma-separated glob patterns to exclude (adds to defaults)")
	rootCmd.Flags().StringSliceVarP(&includePatterns, "include-pattern", "i", []string{}, "Comma-separated glob patterns to include (overrides excludes)")
	rootCmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch to clone and ingest (if source is a Git URL)")
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format: text or json")
}
