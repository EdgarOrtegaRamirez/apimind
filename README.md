# Apimind — API Version Compatibility Analyzer

A CLI tool that compares OpenAPI/Swagger API specifications to detect breaking changes, deprecations, additions, and generate migration reports.

## What It Does

Apimind helps API teams maintain backward compatibility by analyzing differences between API versions. It identifies:

- **Breaking changes**: Removed endpoints, changed required fields, modified response types
- **Deprecations**: Marked deprecated fields, endpoints, or parameters
- **Additions**: New endpoints, optional fields, new operations
- **Modifications**: Changed types, altered validation rules, modified descriptions

## Quick Start

```bash
# Compare two OpenAPI specs
apimind diff spec-v1.json spec-v2.json

# Compare specs from URLs
apimind diff https://api.example.com/v1/openapi.json https://api.example.com/v2/openapi.json

# Generate a migration guide
apimind diff spec-v1.json spec-v2.json --format migration

# Output as JSON for CI/CD
apimind diff spec-v1.json spec-v2.json --format json --output report.json
```

## Installation

```bash
go install github.com/EdgarOrtegaRamirez/apimind/cmd/apimind@latest
```

## Output Formats

- `text` (default): Human-readable report with severity indicators
- `json`: Machine-readable JSON report
- `markdown`: Formatted markdown report
- `migration`: Step-by-step migration guide

## Severity Levels

- **CRITICAL**: Breaking change that will cause client failures
- **WARNING**: Potential compatibility issue
- **INFO**: Non-breaking change worth noting
- **DEPRECATED**: Feature marked for removal

## Example

```
📊 API Compatibility Report
────────────────────────────────────────

Version: v1.0.0 → v2.0.0

🔴 CRITICAL (3)
  • DELETE /api/v2/users/{id} — endpoint removed
  • PATCH /api/v2/users/{id} — required field 'email' added
  • GET /api/v2/products — response type changed: string → object

🟡 WARNING (2)
  • POST /api/v2/orders — parameter 'currency' type changed: string → enum
  • GET /api/v2/users/{id} — response field 'avatar_url' removed

🟢 ADDITIONS (5)
  • PUT /api/v2/users/{id}/avatar — new endpoint
  • GET /api/v2/health — new endpoint
  • POST /api/v2/orders — new optional field 'gift_message'
  • GET /api/v2/users/{id} — new response field 'preferences'
  • GET /api/v2/products — new response field 'tags'

🔵 DEPRECATED (1)
  • GET /api/v1/users — endpoint deprecated, use /api/v2/users

Summary: 3 breaking, 2 warnings, 5 additions, 1 deprecated
```

## CI/CD Integration

```yaml
- name: Check API Compatibility
  run: |
    apimind diff old-spec.json new-spec.json --format json --min-severity warning
    if [ $? -ne 0 ]; then
      echo "Breaking changes detected!"
      exit 1
    fi
```

## Architecture

- `cmd/apimind/` — CLI entry point with Cobra commands
- `internal/loader/` — OpenAPI spec loader (file and URL)
- `internal/comparator/` — Diff engine comparing two specs
- `internal/reporter/` — Output formatters (text, JSON, markdown, migration)
- `internal/model/` — Data structures for API elements

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/getkin/getkin-openapi/openapi3` — OpenAPI 3.x parsing
- Standard library for file I/O and HTTP

## License

MIT