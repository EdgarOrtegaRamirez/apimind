// Package tests provides end-to-end tests for apimind.
package tests

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/EdgarOrtegaRamirez/apimind/internal/comparator"
	"github.com/EdgarOrtegaRamirez/apimind/internal/loader"
	"github.com/EdgarOrtegaRamirez/apimind/internal/model"
	"github.com/EdgarOrtegaRamirez/apimind/internal/reporter"
)

var fixturesDir string

func init() {
	_, file, _, ok := runtime.Caller(0)
	if ok {
		fixturesDir = filepath.Dir(file)
	}
}

func TestCompareBasic(t *testing.T) {
	l := loader.New()

	oldSpec, err := l.Load(filepath.Join(fixturesDir, "spec-v1.yaml"))
	if err != nil {
		t.Fatalf("failed to load old spec: %v", err)
	}

	newSpec, err := l.Load(filepath.Join(fixturesDir, "spec-v2.yaml"))
	if err != nil {
		t.Fatalf("failed to load new spec: %v", err)
	}

	c := comparator.New()
	diff := c.Compare(oldSpec, newSpec)

	if diff == nil {
		t.Fatal("expected non-nil diff")
	}

	if diff.OldVersion != "1.0.0" {
		t.Errorf("expected old version 1.0.0, got %s", diff.OldVersion)
	}
	if diff.NewVersion != "2.0.0" {
		t.Errorf("expected new version 2.0.0, got %s", diff.NewVersion)
	}

	if diff.Total == 0 {
		t.Error("expected changes to be detected, but got 0")
	}
}

func TestDiffHasBreakingChanges(t *testing.T) {
	l := loader.New()

	oldSpec, err := l.Load(filepath.Join(fixturesDir, "spec-v1.yaml"))
	if err != nil {
		t.Fatalf("failed to load old spec: %v", err)
	}

	newSpec, err := l.Load(filepath.Join(fixturesDir, "spec-v2.yaml"))
	if err != nil {
		t.Fatalf("failed to load new spec: %v", err)
	}

	c := comparator.New()
	diff := c.Compare(oldSpec, newSpec)

	if diff.Breaking < 1 {
		t.Errorf("expected at least 1 breaking change, got %d", diff.Breaking)
	}
}

func TestDiffHasAdditions(t *testing.T) {
	l := loader.New()

	oldSpec, err := l.Load(filepath.Join(fixturesDir, "spec-v1.yaml"))
	if err != nil {
		t.Fatalf("failed to load old spec: %v", err)
	}

	newSpec, err := l.Load(filepath.Join(fixturesDir, "spec-v2.yaml"))
	if err != nil {
		t.Fatalf("failed to load new spec: %v", err)
	}

	c := comparator.New()
	diff := c.Compare(oldSpec, newSpec)

	if diff.Added < 1 {
		t.Errorf("expected additions, got %d", diff.Added)
	}
}

func TestReporterText(t *testing.T) {
	diff := &model.APIDiff{
		OldVersion: "1.0.0",
		NewVersion: "2.0.0",
		OldTitle:   "Old API",
		NewTitle:   "New API",
		Changes: []model.DiffChange{
			{
				Severity:  model.Critical,
				Type:      model.Removed,
				Path:      "/api/users",
				Operation: "GET",
				Detail:    "endpoint removed",
			},
		},
	}

	r := reporter.New()
	output, err := r.Generate(diff, reporter.FormatText)
	if err != nil {
		t.Fatalf("text generation failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("expected non-empty text output")
	}
}

func TestReporterJSON(t *testing.T) {
	diff := &model.APIDiff{
		OldVersion: "1.0.0",
		NewVersion: "2.0.0",
		Changes:    []model.DiffChange{},
	}

	r := reporter.New()
	output, err := r.Generate(diff, reporter.FormatJSON)
	if err != nil {
		t.Fatalf("JSON generation failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("expected non-empty JSON output")
	}
}

func TestReporterMarkdown(t *testing.T) {
	diff := &model.APIDiff{
		OldVersion: "1.0.0",
		NewVersion: "2.0.0",
		Changes: []model.DiffChange{
			{
				Severity:  model.Critical,
				Type:      model.Removed,
				Path:      "/api/users",
				Operation: "GET",
				Detail:    "endpoint removed",
			},
		},
	}

	r := reporter.New()
	output, err := r.Generate(diff, reporter.FormatMarkdown)
	if err != nil {
		t.Fatalf("markdown generation failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("expected non-empty markdown output")
	}
}