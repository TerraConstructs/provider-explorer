package ui_test

import (
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/gkampitakis/go-snaps/snaps"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func TestUI_NavigationViewBaseline(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:   "navigation_standard_120x30",
			width:  120,
			height: 30,
		},
		{
			name:   "navigation_compact_80x24",
			width:  80,
			height: 24,
		},
		{
			name:   "navigation_wide_160x40",
			width:  160,
			height: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the TUI with test size
			m := ui.NewModelWithSchemas(ps, tt.width, tt.height)

			tm := teatest.NewTestModel(t, m,
				teatest.WithInitialTermSize(tt.width, tt.height),
			)

			// Let it render the navigation view
			time.Sleep(150 * time.Millisecond)

			// Get the output
			output := tm.Output()
			buf := make([]byte, 8192)
			n, _ := output.Read(buf)
			result := string(buf[:n])

			// Take snapshot
			snaps.MatchSnapshot(t, result)

			// Quit the test
			tm.Quit()
			tm.WaitFinished(t, teatest.WithFinalTimeout(1*time.Second))
		})
	}
}

func TestUI_TreeViewBaseline(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:   "tree_view_standard_120x30",
			width:  120,
			height: 30,
		},
		{
			name:   "tree_view_compact_80x24",
			width:  80,
			height: 24,
		},
		{
			name:   "tree_view_wide_160x40",
			width:  160,
			height: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the TUI with test size
			m := ui.NewModelWithSchemas(ps, tt.width, tt.height)

			tm := teatest.NewTestModel(t, m,
				teatest.WithInitialTermSize(tt.width, tt.height),
			)

			// Navigate to tree view by simulating user interaction
			// 1. Let initial view render
			time.Sleep(100 * time.Millisecond)

			// 2. Select provider (should already be selected)
			tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
			time.Sleep(50 * time.Millisecond)

			// 3. Select resource type (should already be on resources)
			tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
			time.Sleep(50 * time.Millisecond)

			// 4. Select entity (should already be on first resource)
			tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
			time.Sleep(100 * time.Millisecond)

			// Get the tree view output
			output := tm.Output()
			buf := make([]byte, 8192)
			n, _ := output.Read(buf)
			result := string(buf[:n])

			// Take snapshot
			snaps.MatchSnapshot(t, result)

			// Quit the test
			tm.Quit()
			tm.WaitFinished(t, teatest.WithFinalTimeout(1*time.Second))
		})
	}
}

func TestUI_TreeViewWidthComparison(t *testing.T) {
	// This test specifically captures the width jumping issue
	// by testing different entity selections
	lipgloss.SetColorProfile(0)

	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Test with standard size
	width, height := 120, 30
	m := ui.NewModelWithSchemas(ps, width, height)

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(width, height),
	)

	// Navigate to tree view
	time.Sleep(100 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // Enter to select provider
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // Enter to select resource type
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // Enter to select first entity
	time.Sleep(100 * time.Millisecond)

	// Capture first entity state
	output1 := tm.Output()
	buf1 := make([]byte, 8192)
	n1, _ := output1.Read(buf1)
	result1 := string(buf1[:n1])

	// Navigate to next entity to see width change
	tm.Send(tea.KeyMsg{Type: tea.KeyTab}) // Tab to entities pane
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyDown}) // Down arrow to next entity
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // Enter to select next entity
	time.Sleep(100 * time.Millisecond)

	// Capture second entity state
	output2 := tm.Output()
	buf2 := make([]byte, 8192)
	n2, _ := output2.Read(buf2)
	result2 := string(buf2[:n2])

	// Take snapshots of both states - they will use the test name automatically
	t.Run("tree_view_entity1", func(t *testing.T) {
		snaps.MatchSnapshot(t, result1)
	})
	t.Run("tree_view_entity2", func(t *testing.T) {
		snaps.MatchSnapshot(t, result2)
	})

	// Quit the test
	tm.Quit()
	tm.WaitFinished(t, teatest.WithFinalTimeout(1*time.Second))
}
