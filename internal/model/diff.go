// Package model defines data structures for API diff analysis.
package model

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// Severity represents the severity of a diff change.
type Severity string

const (
	Critical   Severity = "CRITICAL"
	Warning    Severity = "WARNING"
	Info       Severity = "INFO"
	Deprecated Severity = "DEPRECATED"
)

// ChangeType represents the kind of change detected.
type ChangeType string

const (
	Added    ChangeType = "ADDED"
	Removed  ChangeType = "REMOVED"
	Modified ChangeType = "MODIFIED"
	Moved    ChangeType = "MOVED"
)

// DiffChange represents a single detected change between two API versions.
type DiffChange struct {
	Severity   Severity   `json:"severity"`
	Type       ChangeType `json:"type"`
	Path       string     `json:"path"`
	Operation  string     `json:"operation,omitempty"`
	Detail     string     `json:"detail"`
	Suggestion string     `json:"suggestion,omitempty"`
}

// APIDiff is the result of comparing two API specifications.
type APIDiff struct {
	OldVersion string       `json:"old_version"`
	NewVersion string       `json:"new_version"`
	OldTitle   string       `json:"old_title"`
	NewTitle   string       `json:"new_title"`
	Changes    []DiffChange `json:"changes"`
	Total      int          `json:"total"`
	Breaking   int          `json:"breaking"`
	Warning    int          `json:"warning"`
	Added      int          `json:"added"`
	Deprecated int          `json:"deprecated"`
}

// EndpointChange represents a change at the endpoint level.
type EndpointChange struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Changed bool   `json:"changed"`
	Removed bool   `json:"removed"`
	Added   bool   `json:"added"`
}

// SchemaDiff tracks differences between two schema objects.
type SchemaDiff struct {
	OldSchema *openapi3.Schema `json:"-"`
	NewSchema *openapi3.Schema `json:"-"`
	Changes   []string         `json:"changes"`
}

// ReportSummary holds a summary of the diff analysis.
type ReportSummary struct {
	TotalChanges    int  `json:"total_changes"`
	HasBreaking     bool `json:"has_breaking"`
	BreakingCount   int  `json:"breaking_count"`
	WARNINGCount    int  `json:"warning_count"`
	AddedCount      int  `json:"added_count"`
	DeprecatedCount int  `json:"deprecated_count"`
}
