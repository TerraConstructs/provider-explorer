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

func Test_HappyPath_ProviderSelect_TypeSelect_EntityBrowse_Export(t *testing.T) {
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

	// 1) Wait until any content renders, then check for providers list
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return len(b) > 0
	}, teatest.WithDuration(3*time.Second))

	// Wait for initial load
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("registry.terraform.io/hashicorp/aws")) &&
			bytes.Contains(b, []byte("press / to filter"))
	}, teatest.WithDuration(5*time.Second))

	// Start filtering with "/" and type "bucket" to narrow to aws_s3_bucket
	tm.Type("/")
	tm.Type("bucket")

	// 5) Press Enter to apply filter and then select aws_s3_bucket
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // apply filter
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // select aws_s3_bucket

	// 6) Wait for tree to show schema, then toggle to Attributes view
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Arguments)"))
	}, teatest.WithDuration(2*time.Second))

	// Toggle to attributes view
	tm.Type("a")

	// Wait for attributes view
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Attributes)"))
	}, teatest.WithDuration(1*time.Second))

	// 7) Select some nodes and export
	tm.Send(tea.KeyMsg{Type: tea.KeySpace}) // select first attribute
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})  // move down
	tm.Send(tea.KeyMsg{Type: tea.KeySpace}) // select second attribute

	// 8) Export (press 'e')
	tm.Type("e")

	// 9) Assert the final output contains our expected HCL output for computed attributes
	teatest.WaitFor(t, tm.FinalOutput(t), func(b []byte) bool {
		return bytes.Contains(b, []byte("Exported HCL")) &&
			(bytes.Contains(b, []byte(`output "`)) || bytes.Contains(b, []byte("Terraform Outputs")))
	}, teatest.WithDuration(2*time.Second))
}

func Test_FilterFocusToggle(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Build the TUI with fixed size for stable output
	m := ui.NewModelWithSchemas(ps, 100, 25)

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(100, 25),
	)

	// Wait for initial load
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("press / to filter"))
	}, teatest.WithDuration(2*time.Second))

	// Test filter activation with "/"
	tm.Type("/") // start filtering

	// Check that filter mode is active
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("filtering"))
	}, teatest.WithDuration(1*time.Second))

	// Press Escape to cancel filtering
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})

	// Check that filter hint appears again
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("press / to filter"))
	}, teatest.WithDuration(1*time.Second))
}

func Test_TreeExpandCollapseSelection(t *testing.T) {
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

	// Navigate to tree view (entities already loaded, just select first one)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // select entity (first one)

	// Wait for tree to appear
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Arguments)"))
	}, teatest.WithDuration(2*time.Second))

	// Test selection
	tm.Send(tea.KeyMsg{Type: tea.KeySpace}) // select current node

	// Check that selection counter appears
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Selected: 1"))
	}, teatest.WithDuration(1*time.Second))

	// Select another node
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})  // move down
	tm.Send(tea.KeyMsg{Type: tea.KeySpace}) // select

	// Check selection counter updates
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Selected: 2"))
	}, teatest.WithDuration(1*time.Second))

	// Test mode toggle
	tm.Type("a") // toggle to attributes

	// Check mode changed
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Attributes)"))
	}, teatest.WithDuration(1*time.Second))
}
