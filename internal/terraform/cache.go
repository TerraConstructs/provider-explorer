package terraform

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/terraconstructs/provider-explorer/internal/schema"
)

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
	
	// Use absolute path and hash it for cache file name
	absPath, err := filepath.Abs(workingDir)
	if err != nil {
		return "", err
	}
	
	// Hash the path for filename
	pathHash := fmt.Sprintf("%x", md5.Sum([]byte(absPath)))
	filename := fmt.Sprintf("provider_schemas_%s.json", pathHash)
	return filepath.Join(cacheDir, filename), nil
}

func ReadProviderSchemaFromCache(workingDir string) (*schema.ProviderSchema, error) {
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

	var providerSchema schema.ProviderSchema
	if err := json.Unmarshal(data, &providerSchema); err != nil {
		return nil, fmt.Errorf("failed to parse cached schema: %w", err)
	}

	return &providerSchema, nil
}

func WriteProviderSchemaToCache(workingDir string, providerSchema *schema.ProviderSchema) error {
	cachePath, err := getCacheFilePath(workingDir)
	if err != nil {
		return err
	}

	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(providerSchema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Provider schemas cached at: %s\n", cachePath)
	return nil
}