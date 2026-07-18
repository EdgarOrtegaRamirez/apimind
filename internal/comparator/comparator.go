// Package comparator analyzes differences between two OpenAPI specs.
package comparator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/EdgarOrtegaRamirez/apimind/internal/model"
	"github.com/getkin/kin-openapi/openapi3"
)

// Comparator compares two OpenAPI specs and produces a diff.
type Comparator struct{}

// New returns a new Comparator.
func New() *Comparator {
	return &Comparator{}
}

// Compare compares two OpenAPI specs and returns a diff.
func (c *Comparator) Compare(old, newSpec *openapi3.T) *model.APIDiff {
	diff := &model.APIDiff{
		OldVersion: old.Info.Version,
		NewVersion: newSpec.Info.Version,
		OldTitle:   old.Info.Title,
		NewTitle:   newSpec.Info.Title,
	}

	// Compare endpoints
	endpointChanges := c.compareEndpoints(old, newSpec)
	diff.Changes = endpointChanges
	diff.Total = len(endpointChanges)
	diff.Breaking = c.countSeverity(diff.Changes, model.Critical)
	diff.Warning = c.countSeverity(diff.Changes, model.Warning)
	diff.Added = c.countType(diff.Changes, model.Added)
	diff.Deprecated = c.countSeverity(diff.Changes, model.Deprecated)

	return diff
}

func (c *Comparator) compareEndpoints(old, newSpec *openapi3.T) []model.DiffChange {
	var changes []model.DiffChange

	oldPaths := old.Paths.Map()
	newPaths := newSpec.Paths.Map()

	// Find removed and modified endpoints
	for path, oldPathItem := range oldPaths {
		newPathItem, exists := newPaths[path]
		if !exists {
			// Endpoint removed — check all methods
			for _, method := range []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"} {
				if getOperation(oldPathItem, method) != nil {
					changes = append(changes, model.DiffChange{
						Severity:  model.Critical,
						Type:      model.Removed,
						Path:      path,
						Operation: strings.ToUpper(method),
						Detail:    fmt.Sprintf("%s endpoint removed", strings.ToUpper(method)),
						Suggestion: "Check if this endpoint was intentionally removed or should be added back",
					})
				}
			}
			continue
		}

		// Compare operations for each method
		for _, method := range []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"} {
			oldOp := getOperation(oldPathItem, method)
			newOp := getOperation(newPathItem, method)
			methodUpper := strings.ToUpper(method)

			if oldOp == nil {
				continue
			}

			if newOp == nil {
				changes = append(changes, model.DiffChange{
					Severity:  model.Critical,
					Type:      model.Removed,
					Path:      path,
					Operation: methodUpper,
					Detail:    fmt.Sprintf("%s operation removed from %s", methodUpper, path),
					Suggestion: "Check if this operation was intentionally removed",
				})
				continue
			}

			// Check for deprecation changes on the operation
			if !oldOp.Deprecated && newOp.Deprecated {
				changes = append(changes, model.DiffChange{
					Severity: model.Deprecated,
					Type:     model.Modified,
					Path:     path,
					Operation: methodUpper,
					Detail:   fmt.Sprintf("%s %s marked as deprecated", methodUpper, path),
				})
			}
			if oldOp.Deprecated && !newOp.Deprecated {
				changes = append(changes, model.DiffChange{
					Severity: model.Deprecated,
					Type:     model.Modified,
					Path:     path,
					Operation: methodUpper,
					Detail:   fmt.Sprintf("deprecated flag removed from %s %s", methodUpper, path),
				})
			}

			// Compare parameters
			changes = append(changes, c.compareParameters(path, methodUpper, oldOp, newOp)...)

			// Compare request body
			changes = append(changes, c.compareRequestBody(path, methodUpper, oldOp, newOp)...)

			// Compare responses
			changes = append(changes, c.compareResponses(path, methodUpper, oldOp, newOp)...)
		}
	}

	// Find added endpoints
	for path, newPathItem := range newPaths {
		if _, exists := oldPaths[path]; !exists {
			for _, method := range []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"} {
				if getOperation(newPathItem, method) != nil {
					changes = append(changes, model.DiffChange{
						Severity:  model.Info,
						Type:      model.Added,
						Path:      path,
						Operation: strings.ToUpper(method),
						Detail:    fmt.Sprintf("%s endpoint added", strings.ToUpper(method)),
					})
				}
			}
		}
	}

	// Sort for deterministic output
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Severity != changes[j].Severity {
			return severityOrderS(changes[i].Severity) < severityOrderS(changes[j].Severity)
		}
		if changes[i].Type != changes[j].Type {
			return severityOrderT(changes[i].Type) < severityOrderT(changes[j].Type)
		}
		return changes[i].Path < changes[j].Path
	})

	return changes
}

func (c *Comparator) compareParameters(path, method string, oldOp, newOp *openapi3.Operation) []model.DiffChange {
	var changes []model.DiffChange

	// Build parameter maps
	oldParams := make(map[string]*openapi3.Parameter)
	for _, pRef := range oldOp.Parameters {
		if pRef != nil && pRef.Value != nil {
			key := paramKey(pRef.Value)
			oldParams[key] = pRef.Value
		}
	}

	newParams := make(map[string]*openapi3.Parameter)
	for _, pRef := range newOp.Parameters {
		if pRef != nil && pRef.Value != nil {
			key := paramKey(pRef.Value)
			newParams[key] = pRef.Value
		}
	}

	// Check for removed, modified, or deprecated parameters
	for key, oldP := range oldParams {
		newP, exists := newParams[key]
		if !exists {
			changes = append(changes, model.DiffChange{
				Severity:  model.Critical,
				Type:      model.Removed,
				Path:      path,
				Operation: method,
				Detail:    fmt.Sprintf("parameter %s.%s removed", oldP.In, oldP.Name),
				Suggestion: "Check if this parameter is still needed",
			})
			continue
		}

		// Check if required changed
		if !oldP.Required && newP.Required {
			changes = append(changes, model.DiffChange{
				Severity:  model.Critical,
				Type:      model.Modified,
				Path:      path,
				Operation: method,
				Detail:    fmt.Sprintf("parameter %s.%s changed from optional to required", oldP.In, oldP.Name),
				Suggestion: "Clients must now provide this parameter",
			})
		}

		// Check if type changed
		oldType := getSchemaTypeName(oldP.Schema)
		newType := getSchemaTypeName(newP.Schema)
		if oldType != newType {
			changes = append(changes, model.DiffChange{
				Severity:  model.Critical,
				Type:      model.Modified,
				Path:      path,
				Operation: method,
				Detail:    fmt.Sprintf("parameter %s.%s type changed from %s to %s", oldP.In, oldP.Name, oldType, newType),
				Suggestion: "Update client code to handle the new type",
			})
		}

		// Check deprecation on parameter
		if !oldP.Deprecated && newP.Deprecated {
			changes = append(changes, model.DiffChange{
				Severity: model.Deprecated,
				Type:     model.Modified,
				Path:     path,
				Operation: method,
				Detail:   fmt.Sprintf("parameter %s.%s deprecated", oldP.In, oldP.Name),
			})
		}
	}

	// Check for added parameters
	for key, newP := range newParams {
		if _, exists := oldParams[key]; !exists {
			severity := model.Info
			if newP.Required {
				severity = model.Warning
			}
			changes = append(changes, model.DiffChange{
				Severity:  severity,
				Type:      model.Added,
				Path:      path,
				Operation: method,
				Detail:    fmt.Sprintf("parameter %s.%s added", newP.In, newP.Name),
			})
		}
	}

	return changes
}

func (c *Comparator) compareRequestBody(path, method string, oldOp, newOp *openapi3.Operation) []model.DiffChange {
	var changes []model.DiffChange

	oldBody := oldOp.RequestBody
	newBody := newOp.RequestBody

	if oldBody != nil && newBody == nil {
		changes = append(changes, model.DiffChange{
			Severity:  model.Critical,
			Type:      model.Removed,
			Path:      path,
			Operation: method,
			Detail:    "request body removed",
			Suggestion: "Check if the endpoint now accepts no body",
		})
		return changes
	}

	if oldBody == nil && newBody != nil {
		changes = append(changes, model.DiffChange{
			Severity:  model.Warning,
			Type:      model.Added,
			Path:      path,
			Operation: method,
			Detail:    "request body added",
		})
		return changes
	}

	if oldBody != nil && newBody != nil {
		oldContent := oldBody.Value.Content
		newContent := newBody.Value.Content

		if len(oldContent) != len(newContent) {
			oldMedia := mediaTypes(oldContent)
			newMedia := mediaTypes(newContent)
			changes = append(changes, model.DiffChange{
				Severity:  model.Critical,
				Type:      model.Modified,
				Path:      path,
				Operation: method,
				Detail:    fmt.Sprintf("request body content type changed: %s → %s", strings.Join(oldMedia, ", "), strings.Join(newMedia, ", ")),
			})
		}

		// Compare JSON schemas
		oldJSON := oldContent["application/json"]
		newJSON := newContent["application/json"]
		if oldJSON != nil && newJSON != nil && oldJSON.Schema != nil && newJSON.Schema != nil {
			changes = append(changes, compareSchemaRefs(path, method, oldJSON.Schema, newJSON.Schema)...)
		}
	}

	return changes
}

func (c *Comparator) compareResponses(path, method string, oldOp, newOp *openapi3.Operation) []model.DiffChange {
	var changes []model.DiffChange

	oldResp := oldOp.Responses
	newResp := newOp.Responses

	if oldResp == nil || newResp == nil {
		return changes
	}

	oldRespMap := oldResp.Map()
	newRespMap := newResp.Map()

	// Check removed response codes
	for code := range oldRespMap {
		if _, exists := newRespMap[code]; !exists {
			changes = append(changes, model.DiffChange{
				Severity:  model.Warning,
				Type:      model.Removed,
				Path:      path,
				Operation: method,
				Detail:    fmt.Sprintf("response status code %s removed", code),
			})
		}
	}

	// Check added response codes
	for code := range newRespMap {
		if _, exists := oldRespMap[code]; !exists {
			changes = append(changes, model.DiffChange{
				Severity:  model.Info,
				Type:      model.Added,
				Path:      path,
				Operation: method,
				Detail:    fmt.Sprintf("response status code %s added", code),
			})
		}
	}

	// Compare response schemas for common codes
	for code, oldRes := range oldRespMap {
		newRes, exists := newRespMap[code]
		if !exists || oldRes == nil || newRes == nil || oldRes.Value == nil || newRes.Value == nil {
			continue
		}

		oldJSON := oldRes.Value.Content["application/json"]
		newJSON := newRes.Value.Content["application/json"]

		if oldJSON != nil && newJSON != nil && oldJSON.Schema != nil && newJSON.Schema != nil {
			changes = append(changes, compareSchemaRefs(path, method, oldJSON.Schema, newJSON.Schema)...)
		}
	}

	return changes
}

// compareSchemaRefs compares two OpenAPI schema refs and returns changes.
func compareSchemaRefs(path, op string, old, new *openapi3.SchemaRef) []model.DiffChange {
	if old == nil || old.Value == nil || new == nil || new.Value == nil {
		return nil
	}
	return compareSchemas(path, op, old.Value, new.Value)
}

// compareSchemas recursively compares two OpenAPI schemas and returns changes.
func compareSchemas(path, op string, old, new *openapi3.Schema) []model.DiffChange {
	var changes []model.DiffChange

	allFields := make(map[string]bool)
	for f := range old.Properties {
		allFields[f] = true
	}
	for f := range new.Properties {
		allFields[f] = true
	}

	for field := range allFields {
		oldHas := old.Properties[field] != nil
		newHas := new.Properties[field] != nil

		newRequired := false
		for _, r := range new.Required {
			if r == field {
				newRequired = true
				break
			}
		}

		if !oldHas && newHas {
			if newRequired {
				changes = append(changes, model.DiffChange{
					Severity: model.Critical,
					Type:     model.Added,
					Path:     fmt.Sprintf("%s.%s", path, field),
					Operation: op,
					Detail:   fmt.Sprintf("required field '%s' added to request/response schema", field),
				})
			} else {
				changes = append(changes, model.DiffChange{
					Severity: model.Info,
					Type:     model.Added,
					Path:     fmt.Sprintf("%s.%s", path, field),
					Operation: op,
					Detail:   fmt.Sprintf("field '%s' added to schema", field),
				})
			}
		} else if oldHas && !newHas {
			changes = append(changes, model.DiffChange{
				Severity: model.Warning,
				Type:     model.Removed,
				Path:     fmt.Sprintf("%s.%s", path, field),
				Operation: op,
				Detail:   fmt.Sprintf("field '%s' removed from schema", field),
				Suggestion: "Update client code to remove references to this field",
			})
		} else if oldHas && newHas {
			// Check type changes
			oldType := getSchemaTypeName(old.Properties[field])
			newType := getSchemaTypeName(new.Properties[field])
			if oldType != newType {
				changes = append(changes, model.DiffChange{
					Severity: model.Critical,
					Type:     model.Modified,
					Path:     fmt.Sprintf("%s.%s", path, field),
					Operation: op,
					Detail:   fmt.Sprintf("field '%s' type changed from %s to %s", field, oldType, newType),
				})
			}
		}
	}

	return changes
}

func getOperation(pi *openapi3.PathItem, method string) *openapi3.Operation {
	switch method {
	case "get":
		return pi.Get
	case "put":
		return pi.Put
	case "post":
		return pi.Post
	case "delete":
		return pi.Delete
	case "options":
		return pi.Options
	case "head":
		return pi.Head
	case "patch":
		return pi.Patch
	case "trace":
		return pi.Trace
	default:
		return nil
	}
}

// Helper functions

func paramKey(p *openapi3.Parameter) string {
	return fmt.Sprintf("%s:%s", p.In, p.Name)
}

func getSchemaTypeName(schema *openapi3.SchemaRef) string {
	if schema == nil || schema.Value == nil {
		return "null"
	}
	if schema.Value.Type != nil {
		return schema.Value.Type.Slice()[0]
	}
	return "object"
}

func getTypeName(schema *openapi3.Schema) string {
	if schema == nil {
		return "null"
	}
	if schema.Type != nil {
		return schema.Type.Slice()[0]
	}
	if len(schema.Properties) > 0 {
		return "object"
	}
	if schema.Items != nil {
		return "array"
	}
	if schema.Enum != nil {
		return "enum"
	}
	if schema.Format != "" {
		return "string:" + schema.Format
	}
	return "string"
}

func mediaTypes(content openapi3.Content) []string {
	var result []string
	for contentType := range content {
		result = append(result, contentType)
	}
	sort.Strings(result)
	return result
}

func (c *Comparator) countSeverity(changes []model.DiffChange, severity model.Severity) int {
	count := 0
	for _, ch := range changes {
		if ch.Severity == severity {
			count++
		}
	}
	return count
}

func (c *Comparator) countType(changes []model.DiffChange, ctype model.ChangeType) int {
	count := 0
	for _, ch := range changes {
		if ch.Type == ctype {
			count++
		}
	}
	return count
}

func severityOrderS(s model.Severity) int {
	switch s {
	case model.Critical:
		return 0
	case model.Warning:
		return 1
	case model.Deprecated:
		return 2
	case model.Info:
		return 3
	default:
		return 4
	}
}

func severityOrderT(t model.ChangeType) int {
	switch t {
	case model.Removed, model.Modified:
		return 0
	case model.Added:
		return 1
	default:
		return 2
	}
}