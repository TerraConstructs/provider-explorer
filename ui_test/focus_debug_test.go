package ui_test

import (
	"path/filepath"
	"testing"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_ProvidersInFullModel(t *testing.T) {
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Create full model
	m := ui.NewModelWithSchemas(ps, 120, 30)
	
	// Test the internal providers component directly from the model
	// This is a bit hacky but let me see if I can access the providers view
	fullView := m.View()
	
	t.Logf("Full model view:\n%s", fullView)
	
	// Create isolated providers component with same setup
	providers := ui.NewProvidersModel(40, 27)  // left width in a 120x30 layout
	providers.SetSchemas(ps)
	providers.Focus()  // match the focus state
	
	isolatedView := providers.View()
	t.Logf("Isolated providers view:\n%s", isolatedView)
}