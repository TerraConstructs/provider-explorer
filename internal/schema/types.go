package schema

import "fmt"

type ProviderTarget struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Source  string `json:"source"`
}

type TerraformConfig struct {
	Provider  map[string]interface{} `json:"provider"`
	Terraform TerraformBlock         `json:"terraform"`
}

type TerraformBlock struct {
	RequiredProviders map[string]RequiredProvider `json:"required_providers"`
}

type RequiredProvider struct {
	Version string `json:"version"`
	Source  string `json:"source"`
}

type ProviderSchema struct {
	FormatVersion    string              `json:"format_version"`
	ProviderSchemas  map[string]Provider `json:"provider_schemas,omitempty"`
	ProviderVersions map[string]string   `json:"provider_versions,omitempty"`
}

type Provider struct {
	Provider          Block                     `json:"provider"`
	ResourceSchemas   map[string]ResourceSchema `json:"resource_schemas,omitempty"`
	DataSourceSchemas map[string]ResourceSchema `json:"data_source_schemas,omitempty"`
}

type ResourceSchema struct {
	Version int   `json:"version"`
	Block   Block `json:"block"`
}

type Block struct {
	Attributes  map[string]Attribute `json:"attributes,omitempty"`
	BlockTypes  map[string]BlockType `json:"block_types,omitempty"`
	Description string               `json:"description,omitempty"`
}

type Attribute struct {
	Type        interface{} `json:"type"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Optional    bool        `json:"optional,omitempty"`
	Computed    bool        `json:"computed,omitempty"`
	Sensitive   bool        `json:"sensitive,omitempty"`
}

type BlockType struct {
	NestingMode string `json:"nesting_mode"`
	Block       Block  `json:"block"`
	MinItems    int    `json:"min_items,omitempty"`
	MaxItems    int    `json:"max_items,omitempty"`
}

type VersionSchema struct {
	ProviderSelections map[string]string `json:"provider_selections"`
}

func GetResourceSchema(providerSchema *ProviderSchema, providerName, resourceName string) (*ResourceSchema, error) {
	provider, exists := providerSchema.ProviderSchemas[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found in schema", providerName)
	}

	resource, exists := provider.ResourceSchemas[resourceName]
	if !exists {
		return nil, fmt.Errorf("resource %s not found in provider %s", resourceName, providerName)
	}

	return &resource, nil
}

func GetDataSourceSchema(providerSchema *ProviderSchema, providerName, dataSourceName string) (*ResourceSchema, error) {
	provider, exists := providerSchema.ProviderSchemas[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found in schema", providerName)
	}

	dataSource, exists := provider.DataSourceSchemas[dataSourceName]
	if !exists {
		return nil, fmt.Errorf("data source %s not found in provider %s", dataSourceName, providerName)
	}

	return &dataSource, nil
}