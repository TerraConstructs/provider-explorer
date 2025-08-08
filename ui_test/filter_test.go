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

func Test_EntityFilteringWorkflow(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture (has aws_instance, aws_s3_bucket)
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Build the TUI with fixed size for stable output
	m := ui.NewModelWithSchemas(ps, 120, 30)

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(120, 30),
	)

	// Wait for entities list to render with both resources and the focus hint
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("aws_instance")) && 
		       bytes.Contains(b, []byte("aws_s3_bucket")) &&
		       bytes.Contains(b, []byte("press / to filter"))
	}, teatest.WithDuration(5*time.Second))

	// Press "/" to start filtering
	tm.Type("/")

	// Wait for filter mode to activate 
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Filter:")) && bytes.Contains(b, []byte("filtering"))
	}, teatest.WithDuration(5*time.Second))

	// Type "bucket" in the filter
	tm.Type("bucket")

	// Give time for the filtering to process 
	time.Sleep(100 * time.Millisecond)

	// Check the filtering results
	output := tm.Output()
	buf := make([]byte, 8192)
	n, _ := output.Read(buf)
	result := string(buf[:n])

	// Verify filtering worked
	hasFilterText := bytes.Contains(buf[:n], []byte("Filter: ")) && bytes.Contains(buf[:n], []byte("bucket"))
	hasInstance := bytes.Contains(buf[:n], []byte("aws_instance"))
	hasBucket := bytes.Contains(buf[:n], []byte("aws_s3_bucket"))
	hasFilterInStatus := bytes.Contains(buf[:n], []byte("filter")) && bytes.Contains(buf[:n], []byte("bucket"))

	t.Logf("Filtering result:\n%s", result)

	if !hasFilterText {
		t.Errorf("FAIL: Filter text 'bucket' not found in filter input")
	}
	if hasInstance {
		t.Errorf("FAIL: aws_instance still visible after filtering")
	}
	if !hasBucket {
		t.Errorf("FAIL: aws_s3_bucket not visible after filtering")
	}
	if !hasFilterInStatus {
		t.Errorf("FAIL: filter not shown in status bar")
	}
	
	tm.Quit()
}

func Test_FilterClearWorkflow(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Build the TUI with fixed size for stable output
	m := ui.NewModelWithSchemas(ps, 120, 30)

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(120, 30),
	)

	// Wait for initial load (simplified condition)
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("press / to filter"))
	}, teatest.WithDuration(5*time.Second))

	// Start filtering 
	tm.Type("/bucket")

	// Press Enter to apply the filter
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for filtering to take effect
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("aws_s3_bucket"))
	}, teatest.WithDuration(5*time.Second))

	// Clear the filter by pressing escape
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})

	// Check that the filter is cleared (look for filter hint)
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("press / to filter"))
	}, teatest.WithDuration(5*time.Second))
}