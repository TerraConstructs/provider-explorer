package schema

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
)

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

// Use official terraform-json ProviderSchemas type
type ProviderSchemas = tfjson.ProviderSchemas
type ProviderSchema = tfjson.ProviderSchema
type Schema = tfjson.Schema
type FunctionSignature = tfjson.FunctionSignature

// Use official terraform-json VersionOutput type
type VersionOutput = tfjson.VersionOutput

func GetResourceSchema(providerSchemas *ProviderSchemas, providerName, resourceName string) (*Schema, error) {
	provider, exists := providerSchemas.Schemas[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found in schema", providerName)
	}

	resource, exists := provider.ResourceSchemas[resourceName]
	if !exists {
		return nil, fmt.Errorf("resource %s not found in provider %s", resourceName, providerName)
	}

	return resource, nil
}

func GetDataSourceSchema(providerSchemas *ProviderSchemas, providerName, dataSourceName string) (*Schema, error) {
	provider, exists := providerSchemas.Schemas[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found in schema", providerName)
	}

	dataSource, exists := provider.DataSourceSchemas[dataSourceName]
	if !exists {
		return nil, fmt.Errorf("data source %s not found in provider %s", dataSourceName, providerName)
	}

	return dataSource, nil
}

func GetEphemeralResourceSchema(providerSchemas *ProviderSchemas, providerName, ephemeralResourceName string) (*Schema, error) {
	provider, exists := providerSchemas.Schemas[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found in schema", providerName)
	}

	ephemeralResource, exists := provider.EphemeralResourceSchemas[ephemeralResourceName]
	if !exists {
		return nil, fmt.Errorf("ephemeral resource %s not found in provider %s", ephemeralResourceName, providerName)
	}

	return ephemeralResource, nil
}

func GetFunctionSchema(providerSchemas *ProviderSchemas, providerName, functionName string) (*FunctionSignature, error) {
	provider, exists := providerSchemas.Schemas[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found in schema", providerName)
	}

	function, exists := provider.Functions[functionName]
	if !exists {
		return nil, fmt.Errorf("function %s not found in provider %s", functionName, providerName)
	}

	return function, nil
}
