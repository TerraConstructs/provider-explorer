package ui_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_DebugProvidersList(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Check the schemas content
	t.Logf("Loaded schemas has %d providers", len(ps.Schemas))
	for name := range ps.Schemas {
		t.Logf("Provider: %s", name)
	}

	// Build the TUI with fixed size for stable output
	m := ui.NewModelWithSchemas(ps, 120, 30)

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(120, 30),
	)

	// Give it some time to initialize
	time.Sleep(100 * time.Millisecond)

	// Get output and log it
	output := tm.Output()
	buf := make([]byte, 4096)
	n, _ := output.Read(buf)

	fullOutput := string(buf[:n])
	t.Logf("Full teatest output:\n%s", fullOutput)

	// Let's try to compare with the non-teatest version
	directView := m.View()
	t.Logf("Direct model.View() output:\n%s", directView)

	tm.Quit()
}
