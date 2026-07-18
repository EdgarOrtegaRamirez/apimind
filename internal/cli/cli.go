// Package cli provides CLI command definitions.
package cli

import (
	"fmt"
	"os"

	"github.com/EdgarOrtegaRamirez/apimind/internal/comparator"
	"github.com/EdgarOrtegaRamirez/apimind/internal/loader"
	"github.com/EdgarOrtegaRamirez/apimind/internal/reporter"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "apimind",
	Short: "API version compatibility analyzer",
	Long: `Apimind compares OpenAPI/Swagger API specifications to detect
breaking changes, deprecations, additions, and generates migration reports.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff [old_spec] [new_spec]",
	Short: "Compare two API specifications",
	Long: `Compare two OpenAPI/Swagger specifications and detect differences.

Examples:
  apimind diff spec-v1.json spec-v2.json
  apimind diff old.yaml new.yaml --format markdown
  apimind diff https://api.example.com/v1/openapi.json ./v2/openapi.json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldPath := args[0]
		newPath := args[1]

		formatStr, _ := cmd.Flags().GetString("format")
		outputPath, _ := cmd.Flags().GetString("output")

		format := reporter.FormatText
		switch formatStr {
		case "json":
			format = reporter.FormatJSON
		case "markdown":
			format = reporter.FormatMarkdown
		case "migration":
			format = reporter.FormatMigration
		case "text":
			format = reporter.FormatText
		}

		result, err := runDiff(oldPath, newPath, format)
		if err != nil {
			return err
		}

		// Handle output
		var output []byte
		if outputPath != "" {
			output = result
			err = os.WriteFile(outputPath, output, 0644)
			if err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Report written to %s\n", outputPath)
		} else {
			fmt.Print(string(result))
		}

		return nil
	},
}

func init() {
	diffCmd.Flags().StringP("format", "f", "text", "Output format: text, json, markdown, migration")
	diffCmd.Flags().StringP("output", "o", "", "Write report to file instead of stdout")
}

func runDiff(oldPath, newPath string, format reporter.Format) ([]byte, error) {
	oldLoader := loader.New()
	newLoader := loader.New()

	oldSpec, err := oldLoader.Load(oldPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load old spec: %w", err)
	}

	newSpec, err := newLoader.Load(newPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load new spec: %w", err)
	}

	comparator := comparator.New()
	diff := comparator.Compare(oldSpec, newSpec)

	reporter := reporter.New()
	return reporter.Generate(diff, format)
}
