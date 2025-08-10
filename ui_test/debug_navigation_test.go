package ui_test

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_DebugNavigationSequence(t *testing.T) {
	lipgloss.SetColorProfile(0)

	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	m := ui.NewModelWithSchemas(ps, 120, 30)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))

	// Step 1: Wait for initial load
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return len(b) > 0
	}, teatest.WithDuration(3*time.Second))

	output := tm.Output()
	buf := make([]byte, 8192)
	n, _ := output.Read(buf)
	t.Logf("=== INITIAL STATE ===\n%s", string(buf[:n]))

	// Step 2: Tab to focus types
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(100 * time.Millisecond)

	n, _ = output.Read(buf)
	t.Logf("=== AFTER TAB (Focus Types) ===\n%s", string(buf[:n]))

	// Step 3: Down to Resources
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)

	n, _ = output.Read(buf)
	t.Logf("=== AFTER DOWN (Should highlight Resources) ===\n%s", string(buf[:n]))

	// Step 4: Enter to select Resources
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(200 * time.Millisecond)

	n, _ = output.Read(buf)
	t.Logf("=== AFTER ENTER (Should show entities list) ===\n%s", string(buf[:n]))

	// Check if both resources are visible
	hasInstance := bytes.Contains(buf[:n], []byte("aws_instance"))
	hasBucket := bytes.Contains(buf[:n], []byte("aws_s3_bucket"))

	t.Logf("Has aws_instance: %v", hasInstance)
	t.Logf("Has aws_s3_bucket: %v", hasBucket)

	tm.Quit()
}
