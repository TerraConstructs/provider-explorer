package ui

import (
	"fmt"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/terraconstructs/provider-explorer/internal/schema"
)

// ResourceSection represents which section of a resource (Arguments or Attributes)
type ResourceSection string

const (
	ArgumentsSection  ResourceSection = "Arguments"
	AttributesSection ResourceSection = "Attributes"
)

// ConvertToHCL converts a resource section to HCL based on the section type
func ConvertToHCL(resourceName string, resourceSchema *schema.Schema, section ResourceSection, providerName string) string {
	switch section {
	case ArgumentsSection:
		return ConvertArgumentsToHCLVariables(resourceSchema)
	case AttributesSection:
		return ConvertAttributesToHCLOutputs(resourceName, resourceSchema, providerName)
	default:
		return "# Unknown section type"
	}
}

// ConvertArgumentsToHCLVariables converts resource arguments to terraform variable blocks
func ConvertArgumentsToHCLVariables(resourceSchema *schema.Schema) string {
	var b strings.Builder

	b.WriteString("# Terraform Variables Generated from Resource Arguments\n\n")

	hasArguments := false
	for name, attr := range resourceSchema.Block.Attributes {
		if attr.Required || attr.Optional {
			hasArguments = true

			// Generate variable block
			b.WriteString(fmt.Sprintf("variable \"%s\" {\n", name))

			// Add type
			hclType := convertTypeToHCLType(attr.AttributeType)
			b.WriteString(fmt.Sprintf("  type = %s\n", hclType))

			// Add description
			description := attr.Description
			if description == "" {
				if attr.Required {
					description = fmt.Sprintf("Required argument for %s", name)
				} else {
					description = fmt.Sprintf("Optional argument for %s", name)
				}
			}
			b.WriteString(fmt.Sprintf("  description = \"%s\"\n", escapeDescription(description)))

			// Add default value for optional arguments
			if !attr.Required {
				b.WriteString("  default = null\n")
			}

			b.WriteString("}\n\n")
		}
	}

	if !hasArguments {
		b.WriteString("# No arguments available for variable conversion\n")
	}

	return b.String()
}

// ConvertAttributesToHCLOutputs converts resource attributes to terraform output blocks
func ConvertAttributesToHCLOutputs(resourceName string, resourceSchema *schema.Schema, providerName string) string {
	var b strings.Builder

	b.WriteString("# Terraform Outputs Generated from Resource Attributes\n\n")

	hasAttributes := false
	for name, attr := range resourceSchema.Block.Attributes {
		if attr.Computed {
			hasAttributes = true

			// Add nested schema as comment if it's a complex type
			if isComplexType(attr.AttributeType) {
				b.WriteString(fmt.Sprintf("# %s structure:\n", name))
				b.WriteString(fmt.Sprintf("# %s\n", renderAttributeTypeComment(attr.AttributeType)))
			}

			// Generate output block
			b.WriteString(fmt.Sprintf("output \"%s\" {\n", name))

			// Generate resource reference
			resourceRef := generateResourceReference(providerName, resourceName, name)
			b.WriteString(fmt.Sprintf("  value = %s\n", resourceRef))

			// Add description
			if attr.Description != "" {
				b.WriteString(fmt.Sprintf("  description = \"%s\"\n", escapeDescription(attr.Description)))
			}

			// Add sensitivity flag if needed
			if attr.Sensitive {
				b.WriteString("  sensitive = true\n")
			}

			b.WriteString("}\n\n")
		}
	}

	if !hasAttributes {
		b.WriteString("# No computed attributes available for output conversion\n")
	}

	return b.String()
}

// ConvertSelectedArgumentsToHCLVariables converts only selected argument attributes into variables.
// It respects hierarchy: a nested attribute is included only if all parent blocks are selected.
func ConvertSelectedArgumentsToHCLVariables(resourceSchema *schema.Schema, selectedPaths [][]string) string {
	if resourceSchema == nil || resourceSchema.Block == nil {
		return "# No arguments available for variable conversion\n"
	}

	// Build a set of selected path keys for ancestor checks
	selSet := make(map[string]struct{})
	for _, p := range selectedPaths {
		selSet[strings.Join(p, ".")] = struct{}{}
	}

	var b strings.Builder
	b.WriteString("# Terraform Variables Generated from Selected Arguments\n\n")

	included := 0
	for _, path := range selectedPaths {
		// Ensure all ancestors are selected
		okAnc := true
		for i := 1; i < len(path); i++ {
			if _, ok := selSet[strings.Join(path[:i], ".")]; !ok {
				okAnc = false
				break
			}
		}
		if !okAnc {
			continue
		}

		// Resolve path to attribute
		if attr, found := resolveAttributeByPath(resourceSchema.Block, path); found {
			// Only include arguments (required/optional, not computed)
			if attr.Required || attr.Optional {
				varName := strings.Join(path, "_")
				b.WriteString(fmt.Sprintf("variable \"%s\" {\n", varName))
				b.WriteString(fmt.Sprintf("  type = %s\n", convertTypeToHCLType(attr.AttributeType)))
				description := attr.Description
				if description == "" {
					if attr.Required {
						description = fmt.Sprintf("Required argument for %s", varName)
					} else {
						description = fmt.Sprintf("Optional argument for %s", varName)
					}
				}
				b.WriteString(fmt.Sprintf("  description = \"%s\"\n", escapeDescription(description)))
				if !attr.Required {
					b.WriteString("  default = null\n")
				}
				b.WriteString("}\n\n")
				included++
			}
		}
	}

	if included == 0 {
		b.WriteString("# No selected arguments available for variable conversion\n")
	}

	return b.String()
}

// ConvertSelectedAttributesToHCLOutputs converts only selected computed attributes into outputs.
// The resource instance name is provided explicitly and the reference path is composed from the selection path.
func ConvertSelectedAttributesToHCLOutputs(resourceName string, resourceSchema *schema.Schema, providerName string, instanceName string, selectedPaths [][]string) string {
	if resourceSchema == nil || resourceSchema.Block == nil {
		return "# No computed attributes available for output conversion\n"
	}

	// Build a set of selected path keys for ancestor checks
	selSet := make(map[string]struct{})
	for _, p := range selectedPaths {
		selSet[strings.Join(p, ".")] = struct{}{}
	}

	var b strings.Builder
	b.WriteString("# Terraform Outputs Generated from Selected Attributes\n\n")

	included := 0
	for _, path := range selectedPaths {
		// Ensure all ancestors are selected
		okAnc := true
		for i := 1; i < len(path); i++ {
			if _, ok := selSet[strings.Join(path[:i], ".")]; !ok {
				okAnc = false
				break
			}
		}
		if !okAnc {
			continue
		}

		if attr, found := resolveAttributeByPath(resourceSchema.Block, path); found {
			if attr.Computed {
				// Name outputs by joining path with underscores for uniqueness
				outName := strings.Join(path, "_")

				// Add nested schema as comment if complex type
				if isComplexType(attr.AttributeType) {
					b.WriteString(fmt.Sprintf("# %s structure:\n", outName))
					b.WriteString(fmt.Sprintf("# %s\n", renderAttributeTypeComment(attr.AttributeType)))
				}

				// Reference uses dot-joined attribute path
				refPath := strings.Join(path, ".")
				resourceRef := fmt.Sprintf("%s.%s.%s", resourceName, instanceName, refPath)

				b.WriteString(fmt.Sprintf("output \"%s\" {\n", outName))
				b.WriteString(fmt.Sprintf("  value = %s\n", resourceRef))
				if attr.Description != "" {
					b.WriteString(fmt.Sprintf("  description = \"%s\"\n", escapeDescription(attr.Description)))
				}
				if attr.Sensitive {
					b.WriteString("  sensitive = true\n")
				}
				b.WriteString("}\n\n")
				included++
			}
		}
	}

	if included == 0 {
		b.WriteString("# No selected computed attributes available for output conversion\n")
	}

	return b.String()
}

// resolveAttributeByPath traverses a schema block hierarchy to find an attribute at the given path.
func resolveAttributeByPath(block *tfjson.SchemaBlock, path []string) (*tfjson.SchemaAttribute, bool) {
	if block == nil {
		return nil, false
	}
	if len(path) == 0 {
		return nil, false
	}
	// Walk through nested blocks until last element
	cur := block
	for i := 0; i < len(path)-1; i++ {
		nb, ok := cur.NestedBlocks[path[i]]
		if !ok || nb == nil || nb.Block == nil {
			return nil, false
		}
		cur = nb.Block
	}
	// Last element should resolve to an attribute
	last := path[len(path)-1]
	attr, ok := cur.Attributes[last]
	if !ok || attr == nil {
		return nil, false
	}
	return attr, true
}

// convertTypeToHCLType converts Terraform schema types to HCL variable types
func convertTypeToHCLType(attrType interface{}) string {
	if attrType == nil {
		return "any"
	}
	// For cty.Type, we'll use a simplified string representation
	// This will need to be enhanced later for better type mapping
	typeStr := fmt.Sprintf("%v", attrType)

	// Basic type mapping for common cases
	if strings.Contains(typeStr, "string") {
		return "string"
	}
	if strings.Contains(typeStr, "number") {
		return "number"
	}
	if strings.Contains(typeStr, "bool") {
		return "bool"
	}
	if strings.Contains(typeStr, "list") {
		return "list(any)"
	}
	if strings.Contains(typeStr, "map") {
		return "map(any)"
	}
	if strings.Contains(typeStr, "set") {
		return "set(any)"
	}

	return "any"
}

// convertObjectTypeToHCL converts object type definitions to HCL object syntax
func convertObjectTypeToHCL(objMap map[string]interface{}) string {
	if len(objMap) == 0 {
		return "object({})"
	}

	var fields []string
	for key, value := range objMap {
		fieldType := convertTypeToHCLType(value)
		fields = append(fields, fmt.Sprintf("%s = %s", key, fieldType))
	}

	return fmt.Sprintf("object({\n    %s\n  })", strings.Join(fields, ",\n    "))
}

// convertTupleTypeToHCL converts tuple type definitions to HCL tuple syntax
func convertTupleTypeToHCL(tupleTypes []interface{}) string {
	if len(tupleTypes) == 0 {
		return "tuple([])"
	}

	var types []string
	for _, t := range tupleTypes {
		types = append(types, convertTypeToHCLType(t))
	}

	return fmt.Sprintf("tuple([%s])", strings.Join(types, ", "))
}

// generateResourceReference creates a terraform resource reference
func generateResourceReference(providerName, resourceName, attributeName string) string {
	// Generate instance name by removing provider prefix and converting to snake_case
	instanceName := generateInstanceName(resourceName)

	// Use provider name in comments if needed, but keep standard resource reference format
	_ = providerName // Used for potential future provider-specific formatting
	return fmt.Sprintf("%s.%s.%s", resourceName, instanceName, attributeName)
}

// generateInstanceName creates a reasonable instance name from resource name
func generateInstanceName(resourceName string) string {
	// Remove common prefixes and convert to a simple name
	name := resourceName

	// Remove provider prefixes (aws_, google_, azurerm_, etc.)
	parts := strings.Split(name, "_")
	if len(parts) > 1 {
		// Use the last meaningful part
		name = parts[len(parts)-1]
	}

	// Use "example" as a safe default instance name
	if name == "" || name == resourceName {
		return "example"
	}

	return name
}

// isComplexType checks if an attribute type is complex (nested)
func isComplexType(attrType interface{}) bool {
	if attrType == nil {
		return false
	}
	// For cty.Type, check if it's a complex type
	typeStr := fmt.Sprintf("%v", attrType)
	return strings.Contains(typeStr, "object") || strings.Contains(typeStr, "tuple")
}

// renderAttributeTypeComment renders type information for comments
func renderAttributeTypeComment(attrType interface{}) string {
	if attrType == nil {
		return "unknown"
	}
	// For cty.Type, use its string representation
	return fmt.Sprintf("%v", attrType)
}

// escapeDescription escapes quotes and special characters in descriptions
func escapeDescription(desc string) string {
	// Replace quotes and newlines
	desc = strings.ReplaceAll(desc, "\"", "\\\"")
	desc = strings.ReplaceAll(desc, "\n", "\\n")
	desc = strings.ReplaceAll(desc, "\r", "\\r")
	desc = strings.ReplaceAll(desc, "\t", "\\t")
	return desc
}
