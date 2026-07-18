# Security

Apimind is a read-only tool that analyzes OpenAPI specification files. It does not make network requests to APIs — it only parses specification files.

## Security Considerations

- **Input validation**: All file paths and URLs are validated before use
- **No execution**: The tool only reads and parses specification files — never executes code from specs
- **No data collection**: No telemetry, metrics, or data is sent anywhere
- **Local only**: All analysis happens locally on the machine

## Reporting Security Issues

If you discover a security vulnerability, please email security@edgarortega.dev or open a private GitHub issue.

## Dependencies

Apimind uses minimal dependencies. All are auditable Go packages:

- `github.com/spf13/cobra` — CLI framework (well-maintained, widely used)
- `github.com/getkin/getkin-openapi/openapi3` — OpenAPI 3.x parser (well-maintained, widely used)