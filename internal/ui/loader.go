package ui

import (
	"encoding/json"
	"io"
	"os"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/terraconstructs/provider-explorer/internal/terraform"
)

// LoadProvidersSchemaFromFile loads provider schemas from a JSON file
// This is used by tests to bypass the CLI and load fixtures
func LoadProvidersSchemaFromFile(path string) (*tfjson.ProviderSchemas, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadProvidersSchemaFromReader(f)
}

// LoadProvidersSchemaFromReader loads provider schemas from a reader
func LoadProvidersSchemaFromReader(r io.Reader) (*tfjson.ProviderSchemas, error) {
	var ps tfjson.ProviderSchemas
	dec := json.NewDecoder(r)
	if err := dec.Decode(&ps); err != nil {
		return nil, err
	}
	// NOTE: we intentionally skip ps.Validate() to keep fixtures tiny/flexible.
	return &ps, nil
}

// NewModelWithSchemas creates a model pre-loaded with schemas (for testing)
func NewModelWithSchemas(schemas *tfjson.ProviderSchemas, width, height int) Model {
	m := NewModel(width, height)
	// Prevent background Init from overwriting injected fixtures during tests
	m.disableAutoLoad = true

	// Set tool info first
	m.toolInfo = terraform.TerraformInfo{
		Binary:   "terraform",
		Tool:     "terraform",
		Registry: "registry.terraform.io",
	}
	m.version = "1.10.5"

	// Update all components with schemas and tool info
	m.schemas = schemas
	m.providers.SetSchemas(schemas)
	m.types.SetToolInfo(m.toolInfo, m.version)
	m.status.SetToolInfo(m.toolInfo, m.version)

	// For testing, pre-select the first provider and set up types counts
	if len(schemas.Schemas) > 0 {
		// Get first provider
		var firstProviderName string
		var firstProviderSchema *tfjson.ProviderSchema
		for name, schema := range schemas.Schemas {
			firstProviderName = name
			firstProviderSchema = schema
			break
		}

		// Set up the types counts as if provider was selected
		m.selectedProvider = firstProviderName
		m.types.SetCounts(
			len(firstProviderSchema.DataSourceSchemas),
			len(firstProviderSchema.ResourceSchemas),
			len(firstProviderSchema.EphemeralResourceSchemas),
			len(firstProviderSchema.Functions),
		)

		// Update status with provider selection
		m.status.SetProvider(firstProviderName)

		// For easier testing, also pre-populate entities with Resources type
		m.selectedType = ResourcesType
		m.entities.SetProvider(firstProviderName, firstProviderSchema)
		m.entities.SetType(ResourcesType)
		m.status.SetResourceType("Resources")

		// Set the stage to EntityBrowse since we've pre-selected everything
		m.stage = StageEntityBrowse
		m.focus = FocusEntities // Start focused on entities for easier testing
		m.entities.Focus()
	} else {
		// Focus providers to start the normal flow
		m.providers.Focus()
	}

	return m
}
