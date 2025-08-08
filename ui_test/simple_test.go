package ui_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_SimpleUIStart(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Build the TUI with fixed size (make it wider for better layout)
	m := ui.NewModelWithSchemas(ps, 120, 30)

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(120, 30),
	)

	// Just let it run for a moment to render
	time.Sleep(100 * time.Millisecond)

	// Get the output and print it for debugging
	output := tm.Output()
	buf := make([]byte, 4096)
	n, _ := output.Read(buf)
	t.Logf("UI Output:\n%s", string(buf[:n]))

	// Quit the test
	tm.Quit()
}