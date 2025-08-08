package config

import (
	"strings"

	"github.com/terraconstructs/provider-explorer/internal/terraform"
)

type ProviderInfo struct {
	Name    string
	Version string
	Source  string
}

func GetInstalledProviders(dir string) ([]ProviderInfo, error) {
	schemaWithVersion, err := terraform.FetchAllProviderSchemas(dir)
	if err != nil {
		return nil, err
	}

	var providers []ProviderInfo
	for name := range schemaWithVersion.Schemas.Schemas {
		// Extract source from provider name if it contains registry info
		parts := strings.Split(name, "/")
		source := name
		displayName := name

		if len(parts) >= 3 {
			// Format: registry.terraform.io/hashicorp/aws
			source = name
			displayName = parts[len(parts)-1] // Just "aws"
		} else if len(parts) == 2 {
			// Format: hashicorp/aws
			tfInfo := terraform.FindTerraformBinary()
			source = tfInfo.Registry + "/" + name
			displayName = parts[1] // Just "aws"
		}

		version := ""
		if schemaWithVersion.VersionInfo != nil && schemaWithVersion.VersionInfo.ProviderSelections != nil {
			version = schemaWithVersion.VersionInfo.ProviderSelections[name]
		}

		providers = append(providers, ProviderInfo{
			Name:    displayName,
			Version: version,
			Source:  source,
		})
	}

	return providers, nil
}
