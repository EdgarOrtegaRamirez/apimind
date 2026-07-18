// Package reporter generates output reports from API diff results.
package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/apimind/internal/model"
)

// Format represents the output format.
type Format string

const (
	FormatText     Format = "text"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
	FormatMigration Format = "migration"
)

// Reporter generates text, JSON, markdown, or migration reports.
type Reporter struct{}

// New returns a new Reporter.
func New() *Reporter {
	return &Reporter{}
}

// Generate produces a report from the diff.
func (r *Reporter) Generate(diff *model.APIDiff, format Format) ([]byte, error) {
	switch format {
	case FormatJSON:
		return r.toJSON(diff)
	case FormatMarkdown:
		return r.toMarkdown(diff), nil
	case FormatMigration:
		return r.toMigration(diff), nil
	case FormatText:
		fallthrough
	default:
		return r.toText(diff), nil
	}
}

func (r *Reporter) toText(diff *model.APIDiff) []byte {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "📊 API Compatibility Report\n")
	fmt.Fprintf(&buf, "%s\n", strings.Repeat("─", 60))
	fmt.Fprintf(&buf, "Version: %s → %s\n\n", diff.OldVersion, diff.NewVersion)

	if diff.OldTitle != diff.NewTitle {
		fmt.Fprintf(&buf, "Title changed: %s → %s\n\n", diff.OldTitle, diff.NewTitle)
	}

	// Group changes by severity
	for _, severity := range []model.Severity{model.Critical, model.Warning, model.Deprecated, model.Info} {
		var changes []model.DiffChange
		for _, ch := range diff.Changes {
			if ch.Severity == severity {
				changes = append(changes, ch)
			}
		}
		if len(changes) == 0 {
			continue
		}

		label := severityLabel(severity)
		fmt.Fprintf(&buf, "%s %s (%d)\n", severityEmoji(severity), label, len(changes))
		for _, ch := range changes {
			var indent string
			if ch.Type == model.Added {
				indent = "+ "
			} else if ch.Type == model.Removed {
				indent = "- "
			} else {
				indent = "~ "
			}
			detail := ch.Detail
			if ch.Operation != "" {
				detail = fmt.Sprintf("%s %s — %s", ch.Operation, ch.Path, ch.Detail)
			}
			fmt.Fprintf(&buf, "  %s%s\n", indent, detail)
			if ch.Suggestion != "" {
				fmt.Fprintf(&buf, "    💡 %s\n", ch.Suggestion)
			}
		}
		fmt.Fprintln(&buf)
	}

	fmt.Fprintf(&buf, "Summary: %d breaking, %d warnings, %d additions, %d deprecated\n",
		diff.Breaking, diff.Warning, diff.Added, diff.Deprecated)

	return buf.Bytes()
}

func (r *Reporter) toJSON(diff *model.APIDiff) ([]byte, error) {
	return json.MarshalIndent(diff, "", "  ")
}

func (r *Reporter) toMarkdown(diff *model.APIDiff) []byte {
	var buf bytes.Buffer

	buf.WriteString("# API Compatibility Report\n\n")
	buf.WriteString(fmt.Sprintf("**Version:** %s → %s\n\n", diff.OldVersion, diff.NewVersion))

	if diff.OldTitle != diff.NewTitle {
		buf.WriteString(fmt.Sprintf("**Title:** %s → %s\n\n", diff.OldTitle, diff.NewTitle))
	}

	// Summary table
	buf.WriteString("| Severity | Count |\n")
	buf.WriteString("|----------|-------|\n")
	if diff.Breaking > 0 {
		buf.WriteString(fmt.Sprintf("| 🔴 Breaking | %d |\n", diff.Breaking))
	}
	if diff.Warning > 0 {
		buf.WriteString(fmt.Sprintf("| 🟡 Warnings | %d |\n", diff.Warning))
	}
	if diff.Added > 0 {
		buf.WriteString(fmt.Sprintf("| 🟢 Additions | %d |\n", diff.Added))
	}
	if diff.Deprecated > 0 {
		buf.WriteString(fmt.Sprintf("| 🔵 Deprecated | %d |\n", diff.Deprecated))
	}
	buf.WriteString(fmt.Sprintf("| **Total** | **%d** |\n\n", diff.Total))

	for _, severity := range []model.Severity{model.Critical, model.Warning, model.Deprecated, model.Info} {
		changes := filterBySeverity(diff.Changes, severity)
		if len(changes) == 0 {
			continue
		}

		buf.WriteString(fmt.Sprintf("### %s %s\n\n", severityEmoji(severity), severityLabel(severity)))
		for _, ch := range changes {
			buf.WriteString(fmt.Sprintf("- `%s %s` — %s\n", ch.Operation, ch.Path, ch.Detail))
		}
		buf.WriteString("\n")
	}

	return buf.Bytes()
}

func (r *Reporter) toMigration(diff *model.APIDiff) []byte {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("# Migration Guide: %s → %s\n\n", diff.OldVersion, diff.NewVersion))
	buf.WriteString("Follow these steps to migrate your code to the new API version.\n\n")

	step := 1

	// Breaking changes first
	for _, ch := range diff.Changes {
		if ch.Severity != model.Critical {
			continue
		}

		switch ch.Type {
		case model.Removed:
			fmt.Fprintf(&buf, "## Step %d: Remove references to `%s %s`\n\n", step, ch.Operation, ch.Path)
			fmt.Fprintf(&buf, "This endpoint/operation has been removed.\n")
			if ch.Suggestion != "" {
				fmt.Fprintf(&buf, "> 💡 %s\n\n", ch.Suggestion)
			}
			step++
		case model.Modified:
			fmt.Fprintf(&buf, "## Step %d: Update `%s %s`\n\n", step, ch.Operation, ch.Path)
			fmt.Fprintf(&buf, "%s\n", ch.Detail)
			if ch.Suggestion != "" {
				fmt.Fprintf(&buf, "> 💡 %s\n\n", ch.Suggestion)
			}
			step++
		}
	}

	// Additions
	for _, ch := range diff.Changes {
		if ch.Severity != model.Info && ch.Type != model.Added {
			continue
		}
		fmt.Fprintf(&buf, "## Step %d: New — `%s %s`\n\n", step, ch.Operation, ch.Path)
		fmt.Fprintf(&buf, "%s\n\n", ch.Detail)
		step++
	}

	return buf.Bytes()
}

func filterBySeverity(changes []model.DiffChange, severity model.Severity) []model.DiffChange {
	var result []model.DiffChange
	for _, ch := range changes {
		if ch.Severity == severity {
			result = append(result, ch)
		}
	}
	return result
}

func severityEmoji(s model.Severity) string {
	switch s {
	case model.Critical:
		return "🔴"
	case model.Warning:
		return "🟡"
	case model.Deprecated:
		return "🔵"
	case model.Info:
		return "🟢"
	default:
		return "⚪"
	}
}

func severityLabel(s model.Severity) string {
	switch s {
	case model.Critical:
		return "CRITICAL"
	case model.Warning:
		return "WARNING"
	case model.Deprecated:
		return "DEPRECATED"
	case model.Info:
		return "CHANGES"
	default:
		return "UNKNOWN"
	}
}