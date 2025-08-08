package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/terraconstructs/provider-explorer/internal/schema"
)

func FetchAllProviderSchemas(workingDir string) (*schema.ProviderSchema, error) {
	tfInfo := FindTerraformBinary()
	
	if cachedSchema, err := ReadProviderSchemaFromCache(workingDir); err == nil {
		fmt.Fprintf(os.Stderr, "Using cached provider schemas\n")
		return cachedSchema, nil
	}

	fmt.Fprintf(os.Stderr, "Fetching provider schemas...\n")
	schemaCmd := exec.Command(tfInfo.Binary, "providers", "schema", "-json")
	schemaCmd.Dir = workingDir
	schemaCmd.Stderr = os.Stderr
	schemaOutput, err := schemaCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("terraform providers schema failed: %w", err)
	}

	var providerSchema schema.ProviderSchema
	err = json.Unmarshal(schemaOutput, &providerSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema output: %w", err)
	}

	versionCmd := exec.Command(tfInfo.Binary, "version", "-json")
	versionCmd.Dir = workingDir
	versionOutput, err := versionCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to get terraform version info: %v\n", err)
	} else {
		var versionSchema schema.VersionSchema
		if err := json.Unmarshal(versionOutput, &versionSchema); err == nil {
			providerSchema.ProviderVersions = versionSchema.ProviderSelections
		}
	}

	sanitizedSchema := sanitizeProviderSchema(providerSchema)

	if err := WriteProviderSchemaToCache(workingDir, &sanitizedSchema); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache schema: %v\n", err)
	}

	return &sanitizedSchema, nil
}

func sanitizeProviderSchema(providerSchema schema.ProviderSchema) schema.ProviderSchema {
	attributeDoublingFix := func(attr *schema.Attribute) {
		if typeSlice, ok := attr.Type.([]interface{}); ok {
			if len(typeSlice) > 2 {
				attr.Type = typeSlice[:2]
			}
		}
	}

	var sanitizeBlock func(*schema.Block)
	sanitizeBlock = func(block *schema.Block) {
		for name, attr := range block.Attributes {
			attributeDoublingFix(&attr)
			block.Attributes[name] = attr
		}

		for name, blockType := range block.BlockTypes {
			sanitizeBlock(&blockType.Block)
			block.BlockTypes[name] = blockType
		}
	}

	for providerName, provider := range providerSchema.ProviderSchemas {
		sanitizeBlock(&provider.Provider)

		for resourceName, resource := range provider.ResourceSchemas {
			sanitizeBlock(&resource.Block)
			provider.ResourceSchemas[resourceName] = resource
		}

		for dataSourceName, dataSource := range provider.DataSourceSchemas {
			sanitizeBlock(&dataSource.Block)
			provider.DataSourceSchemas[dataSourceName] = dataSource
		}
		
		providerSchema.ProviderSchemas[providerName] = provider
	}

	return providerSchema
}