package terraform

import (
	"os/exec"

	"github.com/Masterminds/semver/v3"
)

type Feature string

const (
	ProviderFunctions  Feature = "provider_functions"
	EphemeralResources Feature = "ephemeral_resources"
)

// FeatureSupport maps tools to their minimum version requirements
var FeatureSupport = map[string]map[Feature]string{
	"terraform": {
		ProviderFunctions:  "1.8.0",
		EphemeralResources: "1.10.0",
	},
	"tofu": {
		ProviderFunctions: "1.7.0",
		// https://github.com/opentofu/opentofu/issues/1996#issuecomment-2592644188
		// EphemeralResources: "1.11.0",
	},
}

type TerraformInfo struct {
	Binary   string
	Tool     string // "terraform" or "tofu"
	Registry string
}

// FindTerraformBinary detects available terraform tools and returns info for the preferred one
func FindTerraformBinary() TerraformInfo {
	return FindTerraformBinaryWithPreference("")
}

// FindTerraformBinaryWithPreference allows specifying a preferred tool ("terraform" or "tofu")
// If preference is empty or the preferred tool isn't available, uses default priority
func FindTerraformBinaryWithPreference(preference string) TerraformInfo {
	// Check what tools are available
	_, terraformPathErr := exec.LookPath("terraform")
	_, tofuPathErr := exec.LookPath("tofu")

	// If user has a preference and it's available, use it
	if preference == "terraform" && terraformPathErr == nil {
		return TerraformInfo{
			Binary:   "terraform",
			Tool:     "terraform",
			Registry: "registry.terraform.io",
		}
	}
	if preference == "tofu" && tofuPathErr == nil {
		return TerraformInfo{
			Binary:   "tofu",
			Tool:     "tofu",
			Registry: "registry.opentofu.org",
		}
	}

	// Default priority: prefer terraform if available, otherwise tofu
	if terraformPathErr == nil {
		return TerraformInfo{
			Binary:   "terraform",
			Tool:     "terraform",
			Registry: "registry.terraform.io",
		}
	}

	if tofuPathErr == nil {
		return TerraformInfo{
			Binary:   "tofu",
			Tool:     "tofu",
			Registry: "registry.opentofu.org",
		}
	}

	// Neither tool is available - return terraform as fallback
	// This will likely fail when used, but provides a reasonable default
	return TerraformInfo{
		Binary:   "terraform",
		Tool:     "terraform",
		Registry: "registry.terraform.io",
	}
}

// GetAvailableTools returns a list of available terraform tools
func GetAvailableTools() []string {
	var tools []string

	if _, err := exec.LookPath("terraform"); err == nil {
		tools = append(tools, "terraform")
	}

	if _, err := exec.LookPath("tofu"); err == nil {
		tools = append(tools, "tofu")
	}

	return tools
}

// SupportsFeature checks if the current tool supports a feature at the given version
func (tf TerraformInfo) SupportsFeature(feature Feature, version string) bool {
	// Get the feature requirements for this tool
	toolFeatures, exists := FeatureSupport[tf.Tool]
	if !exists {
		return false
	}

	// Get the minimum version required for this feature
	minVersionStr, exists := toolFeatures[feature]
	if !exists {
		return false
	}

	// Parse versions using semantic versioning
	currentVersion, err := semver.NewVersion(version)
	if err != nil {
		return false
	}

	minVersion, err := semver.NewVersion(minVersionStr)
	if err != nil {
		return false
	}

	// Check if current version meets minimum requirement
	// Note: version 999.0.0 means "not supported"
	if minVersion.Major() >= 999 {
		return false
	}

	return currentVersion.GreaterThanEqual(minVersion)
}
