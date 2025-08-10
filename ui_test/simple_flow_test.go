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

func Test_SimpleFlowWithoutExport(t *testing.T) {
	lipgloss.SetColorProfile(0)

	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	m := ui.NewModelWithSchemas(ps, 120, 30)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))

	// 1) Wait for initial load
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("press / to filter"))
	}, teatest.WithDuration(5*time.Second))

	// 2) Start filtering
	tm.Type("/bucket")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // apply filter

	// 3) Wait for filter to be applied
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("aws_s3_bucket"))
	}, teatest.WithDuration(5*time.Second))

	// 4) Select the entity to go to tree view
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// 5) Wait for tree view
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Arguments)"))
	}, teatest.WithDuration(5*time.Second))

	// 6) Toggle to attributes
	tm.Type("a")

	// 7) Wait for attributes view
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Attributes)"))
	}, teatest.WithDuration(2*time.Second))

	t.Log("✅ Simple flow completed successfully")

	tm.Quit()
}
