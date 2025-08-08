package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/terraconstructs/provider-explorer/internal/schema"
)

// SchemaWithVersionInfo combines provider schemas with version information
type SchemaWithVersionInfo struct {
	Schemas     *schema.ProviderSchemas `json:"schemas"`
	VersionInfo *schema.VersionOutput   `json:"version_info,omitempty"`
	TfInfo      TerraformInfo           `json:"terraform_info"`
}

func FetchAllProviderSchemas(workingDir string) (*SchemaWithVersionInfo, error) {
	tfInfo := FindTerraformBinary()

    if cachedSchema, err := ReadProviderSchemaFromCache(workingDir); err == nil {
        // Always update TfInfo with current binary detection to ensure correct tool priority
        cachedSchema.TfInfo = tfInfo
        return cachedSchema, nil
    }

    // Fetch provider schemas from the tool
    schemaCmd := exec.Command(tfInfo.Binary, "providers", "schema", "-json")
    schemaCmd.Dir = workingDir
    schemaCmd.Stderr = os.Stderr
	schemaOutput, err := schemaCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("terraform providers schema failed: %w", err)
	}

	var providerSchemas schema.ProviderSchemas
	err = json.Unmarshal(schemaOutput, &providerSchemas)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema output: %w", err)
	}

	// Get version information
	var versionInfo *schema.VersionOutput
	versionCmd := exec.Command(tfInfo.Binary, "version", "-json")
	versionCmd.Dir = workingDir
	versionOutput, err := versionCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to get terraform version info: %v\n", err)
	} else {
		var version schema.VersionOutput
		if err := json.Unmarshal(versionOutput, &version); err == nil {
			versionInfo = &version
		}
	}

	// Create the combined schema with version info
	schemaWithVersion := &SchemaWithVersionInfo{
		Schemas:     &providerSchemas,
		VersionInfo: versionInfo,
		TfInfo:      tfInfo,
	}

	if err := WriteProviderSchemaToCache(workingDir, schemaWithVersion); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache schema: %v\n", err)
	}

	return schemaWithVersion, nil
}
