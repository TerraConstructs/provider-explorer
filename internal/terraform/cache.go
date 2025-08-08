package terraform

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// ProviderSpec represents a provider specification for cache key generation
type ProviderSpec struct {
	Source      string
	Constraints string
}

// getLockfileHash reads the .terraform.lock.hcl file and returns its SHA256 hash
func getLockfileHash(workingDir string) (string, error) {
	lockfilePath := filepath.Join(workingDir, ".terraform.lock.hcl")

	data, err := os.ReadFile(lockfilePath)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

// getProviderSpecs extracts provider requirements from terraform configuration
func getProviderSpecs(workingDir string) ([]ProviderSpec, error) {
	module, diags := tfconfig.LoadModule(workingDir)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to load terraform module: %v", diags)
	}

	var specs []ProviderSpec
	for _, provider := range module.RequiredProviders {
		spec := ProviderSpec{
			Source:      provider.Source,
			Constraints: strings.Join(provider.VersionConstraints, ","),
		}
		specs = append(specs, spec)
	}

	// Sort specs for consistent ordering
	sort.Slice(specs, func(i, j int) bool {
		return specs[i].Source < specs[j].Source
	})

	return specs, nil
}

// generateProviderSpecHash creates a hash from provider specifications
func generateProviderSpecHash(specs []ProviderSpec) string {
	var parts []string
	for _, spec := range specs {
		part := fmt.Sprintf("%s@%s", spec.Source, spec.Constraints)
		parts = append(parts, part)
	}

	combined := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)
}

func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".resource-cache"), nil
}

func getCacheFilePath(workingDir string) (string, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", err
	}

	var cacheKey string

	// Strategy 1: Try to use lockfile hash (most reliable)
	if lockfileHash, err := getLockfileHash(workingDir); err == nil {
		cacheKey = lockfileHash
		// fmt.Fprintf(os.Stderr, "Using lockfile-based cache key\n")
	} else {
		// Strategy 2: Use provider specifications from terraform config
		if specs, err := getProviderSpecs(workingDir); err == nil && len(specs) > 0 {
			cacheKey = generateProviderSpecHash(specs)
			fmt.Fprintf(os.Stderr, "Using provider-spec-based cache key\n")
		}
	}

	filename := fmt.Sprintf("provider_schemas_%s.json", cacheKey)
	return filepath.Join(cacheDir, filename), nil
}

// HasValidProviderCache checks if a valid cache exists for the given working directory
func HasValidProviderCache(workingDir string) bool {
	cachePath, err := getCacheFilePath(workingDir)
	if err != nil {
		return false
	}

	// Check if cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return false
	}

	// Try to read and parse the cache to ensure it's valid
	_, err = ReadProviderSchemaFromCache(workingDir)
	return err == nil
}

func ReadProviderSchemaFromCache(workingDir string) (*SchemaWithVersionInfo, error) {
	cachePath, err := getCacheFilePath(workingDir)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cache file not found: %s", cachePath)
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var schemaWithVersion SchemaWithVersionInfo
	if err := json.Unmarshal(data, &schemaWithVersion); err != nil {
		return nil, fmt.Errorf("failed to parse cached schema: %w", err)
	}

	return &schemaWithVersion, nil
}

func WriteProviderSchemaToCache(workingDir string, schemaWithVersion *SchemaWithVersionInfo) error {
	cachePath, err := getCacheFilePath(workingDir)
	if err != nil {
		return err
	}

	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(schemaWithVersion, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Provider schemas cached at: %s\n", cachePath)
	return nil
}
