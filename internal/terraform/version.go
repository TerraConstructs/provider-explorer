package terraform

import (
	"github.com/terraconstructs/provider-explorer/internal/schema"
)

// VersionSupport contains information about what features are supported by the current version
type VersionSupport struct {
	ProviderFunctions  bool
	EphemeralResources bool
	TerraformVersion   string
	Tool               string
}

// GetVersionSupport analyzes the version information and returns what features are supported
func GetVersionSupport(versionInfo *schema.VersionOutput, tfInfo TerraformInfo) *VersionSupport {
	support := &VersionSupport{
		ProviderFunctions:  false,
		EphemeralResources: false,
		Tool:               tfInfo.Tool, // Always set the tool info
	}

	if versionInfo == nil || versionInfo.Version == "" {
		return support
	}

	support.TerraformVersion = versionInfo.Version

	// Use the feature matrix to determine support
	support.ProviderFunctions = tfInfo.SupportsFeature(ProviderFunctions, versionInfo.Version)
	support.EphemeralResources = tfInfo.SupportsFeature(EphemeralResources, versionInfo.Version)

	return support
}

// GetVersionWarning returns a user-friendly warning message for unsupported features
func GetVersionWarning(feature string, versionSupport *VersionSupport) string {
	if versionSupport == nil {
		return feature + " may not be available (version information unavailable)"
	}

	version := versionSupport.TerraformVersion
	if version == "" {
		version = "unknown"
	}

	tool := versionSupport.Tool
	if tool == "" {
		tool = "terraform"
	}

	var featureKey Feature
	var supported bool

	switch feature {
	case "Provider Functions":
		featureKey = ProviderFunctions
		supported = versionSupport.ProviderFunctions
	case "Ephemeral Resources":
		featureKey = EphemeralResources
		supported = versionSupport.EphemeralResources
	default:
		return ""
	}

	if !supported {
		// Get the minimum version requirements from the feature matrix
		toolFeatures, exists := FeatureSupport[tool]
		if !exists {
			return feature + " support unknown for " + tool
		}

		minVersion, exists := toolFeatures[featureKey]
		if !exists {
			return feature + " not supported by " + tool
		}

		// Check if it's the "not supported" version
		if minVersion == "999.0.0" {
			return feature + " not supported in " + tool
		}

		// Show version requirement
		toolDisplayName := "Terraform"
		if tool == "tofu" {
			toolDisplayName = "OpenTofu"
		}

		return feature + " requires " + toolDisplayName + " >= " + minVersion + " (current: " + version + ")"
	}

	return ""
}
