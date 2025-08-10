package ui_test

import (
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_DebugFilterSteps(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Build the TUI with fixed size
	m := ui.NewModelWithSchemas(ps, 120, 30)

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(120, 30),
	)

	// Initial state
	time.Sleep(100 * time.Millisecond)
	output := tm.Output()
	buf := make([]byte, 4096)
	n, _ := output.Read(buf)
	t.Logf("Initial state:\n%s", string(buf[:n]))

	// Start filtering
	tm.Type("/")
	time.Sleep(50 * time.Millisecond)
	n, _ = output.Read(buf)
	t.Logf("After '/' (filter start):\n%s", string(buf[:n]))

	// Type "inst"
	tm.Type("inst")
	time.Sleep(100 * time.Millisecond)
	n, _ = output.Read(buf)
	t.Logf("After typing 'inst':\n%s", string(buf[:n]))

	// Apply filter with Enter
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)
	n, _ = output.Read(buf)
	t.Logf("After Enter (apply filter):\n%s", string(buf[:n]))

	// Clear filter with Escape
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})
	time.Sleep(100 * time.Millisecond)
	n, _ = output.Read(buf)
	t.Logf("After Escape (clear applied filter):\n%s", string(buf[:n]))

	tm.Quit()
}
