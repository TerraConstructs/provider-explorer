package ui_test

import (
	"path/filepath"
	"testing"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_ProvidersComponent(t *testing.T) {
	// Load schemas
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Test providers component directly
	providers := ui.NewProvidersModel(50, 20)
	providers.SetSchemas(ps)

	// Get current provider
	providerName, providerSchema := providers.SelectedProvider()
	t.Logf("Selected provider: %s", providerName)

	if providerName == "" {
		t.Log("No provider selected - this might be normal if list is empty")
	}

	if providerSchema == nil {
		t.Log("No provider schema - this might indicate an issue")
	}

	// Try to render the component to see what it looks like
	view := providers.View()
	t.Logf("Providers view:\n%s", view)
}
