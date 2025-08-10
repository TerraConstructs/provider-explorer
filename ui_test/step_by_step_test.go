package ui_test

import (
	"path/filepath"
	"testing"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_StepByStepModelConstruction(t *testing.T) {
	// Load schemas
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Create model step by step like NewModel does
	width, height := 100, 20
	leftWidth := width / 3
	rightWidth := width - leftWidth
	topHeight := 8
	bottomHeight := height - topHeight - 3

	// Create components individually
	providers := ui.NewProvidersModel(leftWidth, height-3)
	_ = ui.NewTypesModel(rightWidth, topHeight)
	_ = ui.NewEntitiesModel(rightWidth, bottomHeight/2)
	_ = ui.NewSchemaTreeModel(rightWidth, bottomHeight/2)
	_ = ui.NewStatusBar(width)

	t.Log("Components created")

	// Set schemas on providers
	providers.SetSchemas(ps)
	t.Log("Schemas set on providers")

	// Check if providers has items now
	providerName, providerSchema := providers.SelectedProvider()
	t.Logf("After setting schemas - Selected provider: %s", providerName)

	if providerSchema != nil {
		t.Logf("Provider schema has %d resources", len(providerSchema.ResourceSchemas))
	}

	// Check providers view
	view := providers.View()
	t.Logf("Providers view after schema set:\n%s", view)

	// Now create full model
	m := ui.NewModelWithSchemas(ps, width, height)
	t.Log("Full model created")

	// Check providers in full model
	fullView := m.View()
	t.Logf("Full model view:\n%s", fullView)
}
