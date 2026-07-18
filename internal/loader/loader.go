// Package loader loads OpenAPI specifications from files or URLs.
package loader

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// Loader loads OpenAPI specs from various sources.
type Loader struct{}

// New returns a new Loader.
func New() *Loader {
	return &Loader{}
}

// Load loads an OpenAPI spec from a file path or URL.
func (l *Loader) Load(path string) (*openapi3.T, error) {
	if isURL(path) {
		return l.loadFromURL(path)
	}
	return l.loadFromFile(path)
}

func (l *Loader) loadFromFile(path string) (*openapi3.T, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	spec, err := openapi3.NewLoader().LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec from %s: %w", path, err)
	}

	if err := spec.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("invalid spec at %s: %w", path, err)
	}

	return spec, nil
}

func (l *Loader) loadFromURL(path string) (*openapi3.T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %s: %w", path, err)
	}

	loader := &openapi3.Loader{
		Context: ctx,
	}

	spec, err := loader.LoadFromURI(u)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec from %s: %w", path, err)
	}

	if err := spec.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid spec at %s: %w", path, err)
	}

	return spec, nil
}

func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}